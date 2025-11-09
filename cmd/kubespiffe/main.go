package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/jsnctl/kubespiffe/pkg/k8s"
)

const (
	DefaultTrustDomain = "example.org"
)

func main() {
	ctx := context.Background()
	cs, err := k8s.GetKubernetesClientset()
	if err != nil {
		log.Fatalf("problem with k8s clientset: %v", err)
	}
	kscs, err := k8s.GetKubespiffeClientset()
	if err != nil {
		log.Fatalf("problem with kubespiffe clientset: %v", err)
	}
	http.HandleFunc("/v1/svid", func(w http.ResponseWriter, r *http.Request) {
		token := k8s.ExtractBearerToken(r.Header.Get("Authorization"))
		if token == "" {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		jwks, err := k8s.GetKubernetesJWKS(ctx)
		if err != nil {
			slog.Error("problem with JWKS")
			return
		}

		claims, err := k8s.VerifyPSAT(token, jwks)
		if err != nil {
			slog.Error("problem with PSAT", "error", err)
			return
		}

		if err := k8s.AttestPod(ctx, cs, kscs, claims["kubernetes.io"].(map[string]any)); err != nil {
			slog.Info("❌ Pod rejected", "error", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("here's an SVID!"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getTrustDomain() string {
	trustDomain, ok := os.LookupEnv("TRUST_DOMAIN")
	if !ok {
		return DefaultTrustDomain
	}
	return trustDomain
}
