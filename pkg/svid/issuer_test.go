package svid

import (
	"crypto/ecdsa"
	"crypto/x509"
	"testing"

	"github.com/jsnctl/kubespiffe/pkg/apis/kubespiffe/v1alpha1"
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

func TestIssueX509SVID(t *testing.T) {
	issuer, err := NewSVIDIssuer()
	require.NoError(t, err)

	wr := &v1alpha1.WorkloadRegistration{
		Spec: v1alpha1.WorkloadRegistrationSpec{
			SPIFFEID: "spiffe://trusted.org/a/spiffeid",
			SVIDType: "svid",
		},
	}
	bytes, err := issuer.IssueX509SVID(wr)

	require.NoError(t, err)
	assert.NotNil(t, bytes)
	assert.NotNil(t, issuer.svids[wr.Spec.SPIFFEID])
}
