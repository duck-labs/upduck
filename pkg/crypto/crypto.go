package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/duck-labs/upduck/pkg/config"
	"github.com/duck-labs/upduck/pkg/types"
)

func GenerateRSAKeys() (*types.RSAKeysConfig, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA private key: %v", err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKey := &privateKey.PublicKey
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	if err := config.EnsureConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to ensure config directory: %v", err)
	}

	if err := os.WriteFile(config.RSAPrivateKey, privateKeyPEM, 0600); err != nil {
		return nil, fmt.Errorf("failed to write private key file: %v", err)
	}

	if err := os.WriteFile(config.RSAPublicKey, publicKeyPEM, 0644); err != nil {
		return nil, fmt.Errorf("failed to write public key file: %v", err)
	}

	return &types.RSAKeysConfig{
		PrivateKey: string(privateKeyPEM),
		PublicKey:  string(publicKeyPEM),
	}, nil
}

func LoadRSAKeys() (*types.RSAKeysConfig, error) {
	privData, err := os.ReadFile(config.RSAPrivateKey)
	if err != nil {
		return nil, err
	}

	pubData, err := os.ReadFile(config.RSAPublicKey)
	if err != nil {
		return nil, err
	}

	return &types.RSAKeysConfig{
		PrivateKey: string(privData),
		PublicKey:  string(pubData),
	}, nil
}

func GetPublicKeyDigest(publicKey string) string {
	hash := sha256.Sum256([]byte(publicKey))
	return hex.EncodeToString(hash[:])[:16]
}
