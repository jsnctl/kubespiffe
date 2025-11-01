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
	_, err := getKubernetesClientset()
	if err != nil {
		log.Fatalf("problem with k8s clientset")
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

		if err := verifyPSAT(token, jwks); err != nil {
			slog.Error("problem with PSAT: %v", err)
			return
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
