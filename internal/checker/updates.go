package checker

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	goversion "github.com/hashicorp/go-version"
	"github.com/ryan-jan/tfvc/internal/registry"
	"github.com/ryan-jan/tfvc/internal/source"
	"github.com/ryan-jan/tfvc/internal/versions"
)

type Client struct {
	Registry      registry.Client
	GitAuth       transport.AuthMethod
	VersionsCache map[string][]*goversion.Version
}

type Update struct {
	LatestMatchingVersion string
	LatestOverallVersion  string
	LatestMatchingUpdate  string
	LatestOverallUpdate   string
}

func (c *Client) Update(s source.Source, current *goversion.Version, constraints goversion.Constraints, includePrerelease bool) (*Update, error) {
	versions, err := c.Versions(s)
	if err != nil {
		return nil, err
	}
	var out Update
	for _, v := range versions {
		if !includePrerelease && v.Prerelease() != "" {
			continue
		}
		versionString := v.Original()
		out.LatestOverallVersion = versionString
		if current != nil && !v.GreaterThan(current) {
			continue
		}
		out.LatestOverallUpdate = versionString
		if constraints == nil || !constraints.Check(v) {
			continue
		}
		out.LatestMatchingVersion = versionString
		if current != nil {
			out.LatestMatchingUpdate = versionString
		}
	}
	return &out, nil
}

func (c *Client) Versions(s source.Source) ([]*goversion.Version, error) {
	if c.VersionsCache == nil {
		c.VersionsCache = make(map[string][]*goversion.Version, 1)
	}
	if versions, ok := c.VersionsCache[s.URI()]; ok {
		return versions, nil
	}
	switch {
	case s.Git != nil:
		git := s.Git
		if githubToken := os.Getenv("GITHUB_TOKEN"); githubToken != "" {
			c.GitAuth = &githttp.BasicAuth{
				Username: githubToken,
			}
		}
		versions, err := versions.Git(git.Remote, c.GitAuth)
		if err != nil {
			return nil, fmt.Errorf("fetch versions from %q: %w", git.Remote, err)
		}
		c.VersionsCache[s.URI()] = versions
		return versions, nil
	case s.Registry != nil:
		reg := s.Registry
		versions, err := versions.Registry(c.Registry, reg.Hostname, reg.Namespace, reg.Name, reg.Provider)
		if err != nil {
			return nil, fmt.Errorf("fetch versions from registry: %w", err)
		}
		c.VersionsCache[s.URI()] = versions
		return versions, nil
	case s.RegistryProvider != nil:
		reg := s.RegistryProvider
		versions, err := versions.RegistryProvider(c.Registry, reg.Namespace, reg.Name)
		if err != nil {
			return nil, fmt.Errorf("fetch versions from registry: %w", err)
		}
		c.VersionsCache[s.URI()] = versions
		return versions, nil
	case s.Local != nil:
		return nil, nil
	default:
		return nil, source.ErrSourceNotSupported
	}
}
