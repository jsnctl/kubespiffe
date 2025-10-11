package main

import (
	"log/slog"
	"os"
	"time"
)

const (
	DefaultTrustDomain = "example.org"
)

func main() {
	for {
		slog.Info("Running kubespiffed...", "trust_domain", getTrustDomain())
		time.Sleep(5 * time.Second)
	}
}

func getTrustDomain() string {
	trustDomain, ok := os.LookupEnv("TRUST_DOMAIN")
	if !ok {
		return DefaultTrustDomain
	}
	return trustDomain
}
