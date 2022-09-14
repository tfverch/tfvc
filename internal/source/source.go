package source

import (
	"errors"
	"fmt"

	"github.com/ryan-jan/tfvc/internal/regsrc"
)

type Source struct {
	Git              *Git
	Registry         *Registry
	RegistryProvider *Registry
	Local            *string
}

func (s Source) Type() string {
	switch {
	case s.Git != nil:
		return "git"
	case s.Registry != nil:
		return "registry"
	case s.Registry != nil:
		return "registryprovider"
	case s.Local != nil:
		return "local"
	}
	return ""
}

func (s Source) URI() string {
	switch {
	case s.Git != nil:
		return s.Git.Remote
	case s.Registry != nil:
		return s.Registry.Normalized
	case s.RegistryProvider != nil:
		return s.RegistryProvider.Normalized
	case s.Local != nil:
		return *s.Local
	}
	return ""
}

var ErrSourceNotSupported = errors.New("source not supported")

func Parse(raw string) (*Source, error) {
	if module, err := regsrc.ParseModuleSource(raw); err == nil {
		out := &Source{
			Registry: &Registry{
				Hostname:   module.Host().Raw,
				Namespace:  module.RawNamespace,
				Name:       module.RawName,
				Provider:   module.RawProvider,
				Normalized: module.Normalized(),
			},
		}
		return out, nil
	}
	proto, detected, err := detect(raw)
	if err != nil {
		return nil, err
	}
	switch proto {
	case "git":
		git, err := parseGitURL(detected)
		if err != nil {
			return nil, err
		}
		return &Source{Git: git}, nil
	case "file":
		return &Source{Local: &detected}, nil
	default:
		return nil, fmt.Errorf("%w: %v (%v)", ErrSourceNotSupported, proto, raw)
	}
}
