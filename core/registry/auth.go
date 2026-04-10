package registry

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
)

type RegistryAuthConfig struct {
	Username      string `json:"username" toml:"username"`
	Password      string `json:"password" toml:"password"`
	Auth          string `json:"auth" toml:"auth"`
	IdentityToken string `json:"identitytoken" toml:"identitytoken"`
}

type RegistriesConfig map[string]RegistryAuthConfig

// Get implements authn.Helper.
func (a RegistriesConfig) Get(host string) (string, string, error) {
	if a == nil {
		return "", "", nil
	}

	auth := a.GetAuthConfig(host)

	if auth == nil {
		return "", "", nil
	}

	if auth.Username != "" {
		return auth.Username, auth.Password, nil
	}
	if auth.IdentityToken != "" {
		return "", auth.IdentityToken, nil
	}
	if auth.Auth != "" {
		decLen := base64.StdEncoding.DecodedLen(len(auth.Auth))
		decoded := make([]byte, decLen)
		_, err := base64.StdEncoding.Decode(decoded, []byte(auth.Auth))
		if err != nil {
			return "", "", err
		}
		user, passwd, ok := strings.Cut(string(decoded), ":")
		if !ok {
			return "", "", fmt.Errorf("invalid decoded auth: %q", decoded)
		}
		return user, strings.Trim(passwd, "\x00"), nil
	}
	return "", "", nil
}

var _ authn.Helper = (RegistriesConfig)(nil)

func (a RegistriesConfig) GetAuthConfig(registry string) *RegistryAuthConfig {
	if a == nil {
		return nil
	}

	auth, ok := a[registry]
	if !ok {
		return nil
	}
	return &auth
}

func NewKeychain(registries RegistriesConfig) authn.Keychain {
	return authn.NewKeychainFromHelper(registries)
}
