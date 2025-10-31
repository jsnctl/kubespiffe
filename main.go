package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultTrustDomain = "example.org"
)

func main() {
	cs, err := getKubernetesClientset()
	if err != nil {
		log.Fatalf("problem with k8s clientset")
	}
	http.HandleFunc("/v1/svid", func(w http.ResponseWriter, r *http.Request) {
		token := extractBearer(r.Header.Get("Authorization"))
		if token == "" {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		review := &authv1.TokenReview{
			Spec: authv1.TokenReviewSpec{Token: token},
		}

		resp, err := cs.AuthenticationV1().TokenReviews().Create(
			context.Background(), review, metav1.CreateOptions{},
		)
		if err != nil || !resp.Status.Authenticated {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		slog.Info("issuing svid...")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
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
