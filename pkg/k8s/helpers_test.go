package k8s

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"testing"

	"github.com/jsnctl/kubespiffe/pkg/apis/kubespiffe/v1alpha1"
	kubespiffefake "github.com/jsnctl/kubespiffe/pkg/generated/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_extractBearer(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "valid bearer token",
			header: "Bearer i-am-a-bearer-token",
			want:   "i-am-a-bearer-token",
		},
		{
			name:   "missing prefix",
			header: "i-might-be-a-bearer-token-but-i-have-no-Bearer-before-me",
			want:   "",
		},
		{
			name:   "empty header",
			header: "",
			want:   "",
		},
		{
			name:   "prefix only",
			header: "Bearer ",
			want:   "",
		},
		{
			name:   "case sensitive prefix",
			header: "bearer i-could-be-a-bearer-token-but-the-guy-before-me-ruined-it",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractBearerToken(tt.header)
			assert.Equal(t, got, tt.want)
		})
	}
}

func mockRSAJWK() map[string]any {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil
	}

	nStr := base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes())

	eBytes := big.NewInt(int64(privateKey.PublicKey.E)).Bytes()
	eStr := base64.RawURLEncoding.EncodeToString(eBytes)

	return map[string]any{
		"kty": "RSA",
		"n":   nStr,
		"e":   eStr,
	}
}

func Test_jwkToPublicKey(t *testing.T) {
	validJWK := mockRSAJWK()

	tests := []struct {
		name    string
		jwk     map[string]any
		wantErr bool
	}{
		{
			name:    "valid JWK",
			jwk:     validJWK,
			wantErr: false,
		},
		{
			name: "missing n",
			jwk: map[string]any{
				"e": validJWK["e"],
			},
			wantErr: true,
		},
		{
			name: "missing e",
			jwk: map[string]any{
				"n": validJWK["n"],
			},
			wantErr: true,
		},
		{
			name: "invalid base64",
			jwk: map[string]any{
				"n": "inv@lid",
				"e": validJWK["e"],
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jwkToPublicKey(tt.jwk)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.IsType(t, &rsa.PublicKey{}, got)
		})
	}
}

func makeClaims(namespace, podName, saName string) map[string]any {
	return map[string]any{
		"iss": "https://kubernetes.default.svc.cluster.local",
		"kubernetes.io": map[string]any{
			"namespace": namespace,
			"pod": map[string]any{
				"name": podName,
				"uid":  "abc-123",
			},
			"serviceaccount": map[string]any{
				"name": saName,
				"uid":  "def-456",
			},
		},
	}
}

func wreg(name, namespace, saName, podName, spiffeID string) *v1alpha1.WorkloadRegistration {
	return &v1alpha1.WorkloadRegistration{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1alpha1.WorkloadRegistrationSpec{
			SPIFFEID: spiffeID,
			SVIDType: "X509",
			Selector: v1alpha1.WorkloadRegistrationSelector{
				Namespace:          namespace,
				ServiceAccountName: saName,
				PodName:            podName,
			},
		},
	}
}

func TestAttestPod(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		registrations []*v1alpha1.WorkloadRegistration
		claims        map[string]any
		wantSPIFFEID  string
		wantErr       bool
	}{
		{
			name: "exact match on all selector fields",
			registrations: []*v1alpha1.WorkloadRegistration{
				wreg("web", "default", "web-sa", "web-pod-abc", "spiffe://example.org/web"),
			},
			claims:       makeClaims("default", "web-pod-abc", "web-sa"),
			wantSPIFFEID: "spiffe://example.org/web",
		},
		{
			name: "matches first of multiple registrations",
			registrations: []*v1alpha1.WorkloadRegistration{
				wreg("other", "default", "other-sa", "other-pod", "spiffe://example.org/other"),
				wreg("web", "default", "web-sa", "web-pod-abc", "spiffe://example.org/web"),
			},
			claims:       makeClaims("default", "web-pod-abc", "web-sa"),
			wantSPIFFEID: "spiffe://example.org/web",
		},
		{
			name: "wildcard pod name matches any pod in namespace",
			registrations: []*v1alpha1.WorkloadRegistration{
				wreg("batch", "jobs", "batch-sa", "", "spiffe://example.org/batch"),
			},
			claims:       makeClaims("jobs", "batch-worker-7f9d2", "batch-sa"),
			wantSPIFFEID: "spiffe://example.org/batch",
		},
		{
			name: "namespace mismatch — no match",
			registrations: []*v1alpha1.WorkloadRegistration{
				wreg("web", "production", "web-sa", "web-pod", "spiffe://example.org/web"),
			},
			claims:  makeClaims("staging", "web-pod", "web-sa"),
			wantErr: true,
		},
		{
			name: "service account mismatch — no match",
			registrations: []*v1alpha1.WorkloadRegistration{
				wreg("web", "default", "restricted-sa", "web-pod", "spiffe://example.org/web"),
			},
			claims:  makeClaims("default", "web-pod", "other-sa"),
			wantErr: true,
		},
		{
			name:          "no registrations — no match",
			registrations: nil,
			claims:        makeClaims("default", "web-pod", "web-sa"),
			wantErr:       true,
		},
		{
			name:    "missing kubernetes.io claims",
			claims:  map[string]any{"iss": "https://kubernetes.default.svc.cluster.local"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjects := make([]runtime.Object, len(tt.registrations))
			for i, r := range tt.registrations {
				runtimeObjects[i] = r
			}
			kscs := kubespiffefake.NewSimpleClientset(runtimeObjects...)

			got, err := AttestPod(ctx, nil, kscs, tt.claims)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantSPIFFEID, got.Spec.SPIFFEID)
		})
	}
}
