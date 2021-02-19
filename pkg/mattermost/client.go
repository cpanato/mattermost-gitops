package mattermost

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

var insecureSignatureAlgorithms = map[x509.SignatureAlgorithm]bool{
	x509.SHA1WithRSA:   true,
	x509.DSAWithSHA1:   true,
	x509.ECDSAWithSHA1: true,
}

func NewAPIv4Client(instanceURL string, allowInsecureSHA1, allowInsecureTLS bool) *model.Client4 {
	client := model.NewAPIv4Client(instanceURL)
	userAgent := fmt.Sprintf("mm-gitops/%s (%s)", "Version", runtime.GOOS)
	client.HttpHeader = map[string]string{"User-Agent": userAgent}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if allowInsecureTLS {
		tlsConfig.MinVersion = tls.VersionTLS10
	}

	if !allowInsecureSHA1 {
		tlsConfig.VerifyPeerCertificate = VerifyCertificates
	}

	client.HttpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return client
}

func InitClientWithCredentials(credentials *Config, allowInsecureSHA1, allowInsecureTLS bool) (*model.Client4, string, error) {
	client := NewAPIv4Client(credentials.InstanceURL, allowInsecureSHA1, allowInsecureTLS)

	client.AuthType = model.HEADER_BEARER
	client.AuthToken = credentials.AuthToken

	_, response := client.GetMe("")
	if response.Error != nil {
		return nil, "", checkInsecureTLSError(response.Error, allowInsecureTLS)
	}

	return client, response.ServerVersion, nil
}

func VerifyCertificates(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	// loop over certificate chains
	for _, chain := range verifiedChains {
		if isValidChain(chain) {
			return nil
		}
	}
	return fmt.Errorf("insecure algorithm found in the certificate chain. Use --insecure-sha1-intermediate flag to ignore. Aborting")
}

func isValidChain(chain []*x509.Certificate) bool {
	// check all certs but the root one
	certs := chain[:len(chain)-1]

	for _, cert := range certs {
		if _, ok := insecureSignatureAlgorithms[cert.SignatureAlgorithm]; ok {
			return false
		}
	}
	return true
}

func checkInsecureTLSError(err *model.AppError, allowInsecureTLS bool) error {
	if (strings.Contains(err.DetailedError, "tls: protocol version not supported") ||
		strings.Contains(err.DetailedError, "tls: server selected unsupported protocol version")) && !allowInsecureTLS {
		return errors.New("won't perform action through an insecure TLS connection. Please add --insecure-tls-version to bypass this check")
	}
	return err
}
