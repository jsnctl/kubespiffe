package main

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jsnctl/kubespiffe/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getKubernetesClientset() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

func getKubespiffeClientset() (*versioned.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return versioned.NewForConfig(cfg)
}

type JWKS struct {
	Keys []map[string]interface{}
}

func getKubernetesJWKS(ctx context.Context) (*JWKS, error) {
	k8sJWKSEndpoint := "https://kubernetes.default.svc/openid/v1/jwks"
	satPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"
	caPath := "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

	token, err := os.ReadFile(satPath)
	if err != nil {
		return nil, fmt.Errorf("reading service account token: %w", err)
	}

	caCertPool, err := loadCertPool(caPath)
	if err != nil {
		return nil, fmt.Errorf("loading CA cert: %w", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: caCertPool},
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k8sJWKSEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+string(token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response: %s", resp.Status)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("decoding JWKS: %w", err)
	}

	return &jwks, nil
}

func loadCertPool(path string) (*x509.CertPool, error) {
	certData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(certData) {
		return nil, fmt.Errorf("failed to append CA certs")
	}
	return pool, nil
}

func extractBearerToken(header string) string {
	hasToken := strings.HasPrefix(header, "Bearer ")
	if !hasToken {
		return ""
	}
	return strings.TrimPrefix(header, "Bearer ")
}

func verifyPSAT(psat string, jwks *JWKS) (map[string]any, error) {
	audience := "kubespiffed"
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	unverifiedPSAT, _, err := parser.ParseUnverified(psat, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("parse unverified token: %w", err)
	}

	header := unverifiedPSAT.Header
	kid, ok := header["kid"].(string)
	if !ok {
		return nil, errors.New("missing kid in token header")
	}

	key, err := findKeyByKID(jwks, kid)
	if err != nil {
		return nil, err
	}

	pubKey, err := jwkToPublicKey(key)
	if err != nil {
		return nil, fmt.Errorf("convert jwk to public key: %w", err)
	}

	verified, err := jwt.Parse(psat, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token signature verification failed: %w", err)
	}

	claims, ok := verified.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	if iss, ok := claims["iss"].(string); !ok || iss != "https://kubernetes.default.svc.cluster.local" {
		return nil, fmt.Errorf("invalid issuer: %v", claims["iss"])
	}
	if aud, ok := claims["aud"].([]interface{}); ok {
		valid := false
		for _, a := range aud {
			if a.(string) == audience {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("token audience %v does not include %q", aud, audience)
		}
	}

	if exp, ok := claims["exp"].(float64); ok && time.Unix(int64(exp), 0).Before(time.Now()) {
		return nil, fmt.Errorf("token expired at %v", time.Unix(int64(exp), 0))
	}

	return claims, nil
}

func findKeyByKID(jwks *JWKS, kid string) (map[string]interface{}, error) {
	for _, key := range jwks.Keys {
		if key["kid"] == kid {
			return key, nil
		}
	}
	return nil, fmt.Errorf("no key found for kid: %s", kid)
}

func jwkToPublicKey(jwk map[string]interface{}) (*rsa.PublicKey, error) {
	// TODO: This can definitely be simplified with the Go JWT
	// library API, this fn was original provided by Gemini and is likely
	// to be over-engineered and potentially buggy
	nStr, okN := jwk["n"].(string)
	eStr, okE := jwk["e"].(string)
	if !okN || !okE {
		return nil, errors.New("missing n or e in jwk")
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("decode n: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("decode e: %w", err)
	}

	var e int
	switch len(eBytes) {
	case 3:
		e = int(binary.BigEndian.Uint32(append([]byte{0}, eBytes...)))
	default:
		e = int(binary.BigEndian.Uint16(eBytes))
	}

	pub := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}
	return pub, nil
}

type KubernetesWorkloadClaims struct {
	Namespace      string             `json:"namespace"`
	Node           KubernetesResource `json:"node"`
	Pod            KubernetesResource `json:"pod"`
	ServiceAccount KubernetesResource `json:"serviceAccount"`
}

type KubernetesResource struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

func attestPod(ctx context.Context, cs *kubernetes.Clientset, kscs *versioned.Clientset, claims map[string]any) error {
	b, err := json.Marshal(claims)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	var c KubernetesWorkloadClaims
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}

	// Quick hacky prune of workload pod name in PSAT claim to test allow/deny policy
	podName := strings.Split(c.Pod.Name, "-")[0]

	_, err = kscs.KubespiffeV1alpha1().WorkloadRegistrations("").Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get registration for %s/%s: %w", c.Namespace, c.Pod.Name, err)
	}

	slog.Info("✅ Pod attested", "pod", c.Pod.Name, "namespace", c.Namespace)
	return nil
}

func checkForLabel(pod *corev1.Pod, key, value string) error {
	val, ok := pod.GetLabels()[key]
	if !ok {
		return fmt.Errorf("pod label does not exist")
	}
	if val != value {
		return fmt.Errorf("pod value does not match expected")
	}
	return nil
}
