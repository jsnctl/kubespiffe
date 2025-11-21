package svid

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/jsnctl/kubespiffe/pkg/apis/kubespiffe/v1alpha1"
)

type SVIDIssuer struct {
	signer crypto.Signer
	caCert *x509.Certificate
	svids  map[string][]byte
}

func NewSVIDIssuer() (*SVIDIssuer, error) {
	caKey, err := createCAKey()
	if err != nil {
		return nil, fmt.Errorf("problem with CA key: %w", err)
	}

	caCert, err := createCACert(caKey)
	if err != nil {
		return nil, fmt.Errorf("problem with CA cert: %w", err)
	}

	svids := make(map[string][]byte)
	return &SVIDIssuer{
		signer: caKey,
		caCert: caCert,
		svids:  svids,
	}, nil
}

func createCAKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func createCACert(key *ecdsa.PrivateKey) (*x509.Certificate, error) {
	format := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "kubespiffe"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		format,
		format,
		&key.PublicKey,
		key,
	)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certBytes)
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
	svidBytes, err := x509.CreateCertificate(rand.Reader, svid, i.caCert, &key.PublicKey, i.signer)
	if err != nil {
		return nil, err
	}

	i.svids[wr.Spec.SPIFFEID] = svidBytes
	return svidBytes, nil
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
