package svid

import (
	"crypto/ecdsa"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSVIDIssuer(t *testing.T) {
	issuer, err := NewSVIDIssuer()
	require.NoError(t, err)
	assert.NotNil(t, issuer)

	assert.IsType(t, &ecdsa.PrivateKey{}, issuer.signer)
	assert.IsType(t, &x509.Certificate{}, issuer.caCert)
	assert.True(t, issuer.caCert.IsCA)
}
