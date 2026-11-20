package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAlipayValidateConfig(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pub, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	cfg := map[string]string{
		"appId": "2021001234567890", "privateKey": string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})),
		"alipayPublicKey": string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pub})),
		"notifyUrl":       "https://pay.example.test/api/v1/payment/webhook/alipay", "returnUrl": "https://pay.example.test/payment/result",
	}
	p, err := NewAlipay("qa", cfg)
	require.NoError(t, err)
	require.NoError(t, p.ValidateConfig())
	cfg["notifyUrl"] = "http://pay.example.test/api/v1/payment/webhook/alipay"
	require.ErrorContains(t, p.ValidateConfig(), "HTTPS")
	cfg["privateKey"] = "invalid-private-key"
	p, err = NewAlipay("qa", cfg)
	require.NoError(t, err)
	require.ErrorContains(t, p.ValidateConfig(), "private key")
}
