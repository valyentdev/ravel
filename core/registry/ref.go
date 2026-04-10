package registry

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
)

type Reference struct {
	Domain     string
	Repository string
	Tag        string
	Digest     string
}

func (r *Reference) HasDigest() bool {
	return r.Digest != ""
}

func (r *Reference) String() string {
	if r.Digest != "" {
		return fmt.Sprintf("%s/%s@%s", r.Domain, r.Repository, r.Digest)
	}

	var tag string
	if r.Tag != "" {
		tag = r.Tag
	} else {
		tag = "latest"
	}

	return fmt.Sprintf("%s/%s:%s", r.Domain, r.Repository, tag)
}

func Parse(ref string) (Reference, error) {
	if t, err := name.NewTag(ref); err == nil {
		return Reference{
			Domain:     t.RegistryStr(),
			Repository: t.RepositoryStr(),
			Tag:        t.TagStr(),
		}, nil
	}
	if d, err := name.NewDigest(ref); err == nil {
		return Reference{
			Domain:     d.RegistryStr(),
			Repository: d.RepositoryStr(),
			Digest:     d.DigestStr(),
		}, nil
	}

	return Reference{}, fmt.Errorf("invalid reference: %s", ref)

}

func CheckImageRef(ctx context.Context, ref Reference, authConfig RegistriesConfig) (string, error) {
	i, err := crane.Pull(ref.String(), crane.WithContext(ctx), crane.WithAuthFromKeychain(NewKeychain(authConfig)))
	if err != nil {
		return "", err
	}

	if ref.HasDigest() {
		return ref.String(), nil
	}

	d, err := i.Digest()
	if err != nil {
		return "", err
	}

	ref.Digest = d.String()

	return ref.String(), nil
}
