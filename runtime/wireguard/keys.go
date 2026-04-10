package wireguard

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

const KeyLength = 32

// Key represents a Wireguard key (either private or public)
type Key [KeyLength]byte

// String returns the base64-encoded string representation of the key
func (k Key) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}

// KeyPair represents a Wireguard private/public keypair
type KeyPair struct {
	PrivateKey Key
	PublicKey  Key
}

// GenerateKeyPair generates a new Wireguard keypair
func GenerateKeyPair() (*KeyPair, error) {
	var privateKey Key

	// Generate random bytes for private key
	if _, err := rand.Read(privateKey[:]); err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Clamp the private key (Wireguard requirement)
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	// Derive public key from private key using Curve25519
	var publicKey Key
	curve25519.ScalarBaseMult((*[32]byte)(&publicKey), (*[32]byte)(&privateKey))

	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

// ParseKey parses a base64-encoded key string
func ParseKey(s string) (Key, error) {
	var key Key
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return key, fmt.Errorf("invalid key encoding: %w", err)
	}
	if len(decoded) != KeyLength {
		return key, fmt.Errorf("invalid key length: expected %d bytes, got %d", KeyLength, len(decoded))
	}
	subtle.ConstantTimeCopy(1, key[:], decoded)
	return key, nil
}
