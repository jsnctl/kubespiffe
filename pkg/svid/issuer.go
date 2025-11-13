package svid

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/url"
	"time"

	"github.com/jsnctl/kubespiffe/pkg/apis/kubespiffe/v1alpha1"
)

type SVIDIssuer struct {
	signer crypto.Signer
	caCert *x509.Certificate
}

func NewSVIDIssuer() *SVIDIssuer {
	caKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "kubespiffe-ca"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caDER, _ := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	caCert, _ := x509.ParseCertificate(caDER)

	return &SVIDIssuer{
		signer: caKey,
		caCert: caCert,
	}
}

func (i *SVIDIssuer) IssueX509SVID(wr *v1alpha1.WorkloadRegistration) ([]byte, error) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	svid := &x509.Certificate{
		SerialNumber:          randomSerial(),
		Subject:               pkixNameFrom(wr.Spec.SPIFFEID),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(5 * time.Minute)),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		URIs:                  []*url.URL{mustParseSPIFFEID(wr.Spec.SPIFFEID)},
		BasicConstraintsValid: true,
	}
	return x509.CreateCertificate(rand.Reader, svid, i.caCert, &key.PublicKey, i.signer)
}

func randomSerial() *big.Int {
	n, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	return n
}

func pkixNameFrom(spiffeID string) pkix.Name {
	return pkix.Name{
		CommonName:   spiffeID,
		Organization: []string{"kubespiffe.io"},
	}
}

func mustParseSPIFFEID(spiffeID string) *url.URL {
	uri, err := url.Parse(spiffeID)
	if err != nil {
		panic(err)
	}
	return uri
}
