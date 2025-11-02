package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
)

const (
	DefaultTrustDomain = "example.org"
)

func main() {
	ctx := context.Background()
	cs, err := getKubernetesClientset()
	if err != nil {
		log.Fatalf("problem with k8s clientset: %v", err)
	}
	kscs, err := getKubespiffeClientset()
	if err != nil {
		log.Fatalf("problem with kubespiffe clientset: %v", err)
	}
	http.HandleFunc("/v1/svid", func(w http.ResponseWriter, r *http.Request) {
		token := extractBearer(r.Header.Get("Authorization"))
		if token == "" {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		jwks, err := getKubernetesJWKS(ctx)
		if err != nil {
			slog.Error("problem with JWKS")
			return
		}

		claims, err := verifyPSAT(token, jwks)
		if err != nil {
			slog.Error("problem with PSAT", "error", err)
			return
		}

		if err := attestPod(ctx, cs, kscs, claims["kubernetes.io"].(map[string]any)); err != nil {
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
