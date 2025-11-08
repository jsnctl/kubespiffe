package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
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
			got := extractBearerToken(tt.header)
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
