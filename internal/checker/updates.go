package checker

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	goversion "github.com/hashicorp/go-version"
	"github.com/tfverch/tfvc/internal/output"
	"github.com/tfverch/tfvc/internal/registry"
	"github.com/tfverch/tfvc/internal/source"
	"github.com/tfverch/tfvc/internal/versions"
)

type Client struct {
	Registry      registry.Client
	GitAuth       transport.AuthMethod
	VersionsCache map[string][]*goversion.Version
}

type Update struct {
	LatestMatchingVersion goversion.Version
	LatestOverallVersion  goversion.Version
	LatestMatchingUpdate  goversion.Version
	LatestOverallUpdate   goversion.Version
}

func updates(p []Parsed, includePrerelease bool, sshPrivKeyPath string, sshPrivKeyPwd string, path string) (output.Updates, error) {
	var updates output.Updates
	for _, parsed := range p {
		client := Client{
			Registry: registry.Client{
				HTTP: http.DefaultClient,
			},
		}
		err := client.config(parsed, sshPrivKeyPath, sshPrivKeyPwd)
		if err != nil {
			return nil, err
		}
		update, err := client.Update(*parsed.Source, parsed.Version, parsed.Constraints, includePrerelease)
		if err != nil {
			return nil, err
		}
		output := output.Update{
			VersionConstraints: parsed.Constraints,
			LatestMatching:     update.LatestMatchingVersion,
			LatestOverall:      update.LatestOverallVersion,
		}
		if parsed.Version != nil {
			output.Version = *parsed.Version
		}
		if parsed.RawProvider != nil {
			output.Type = "provider"
			output.Path = path
			output.Name = parsed.RawProvider.Source
			output.Source = parsed.RawProvider.Source
		}
		if parsed.RawModule != nil {
			output.Type = "module"
			output.Path = parsed.RawModule.Pos.Filename
			output.Name = parsed.RawModule.Name
			output.Source = parsed.RawModule.Source
		}
		output.SetUpdateStatus()
		updates = append(updates, output)
	}
	return updates, nil
}

func (c *Client) config(parsed Parsed, sshPrivKeyPath string, sshPrivKeyPwd string) error {
	source := *parsed.Source
	if source.Git != nil {
		ssh, err := regexp.MatchString("git@", source.Git.Remote)
		if err != nil {
			return fmt.Errorf("Main: error checking git ssh regex %w", err)
		}
		if ssh {
			authSSH, err := gitssh.NewPublicKeysFromFile("git", sshPrivKeyPath, sshPrivKeyPwd)
			if err != nil {
				return fmt.Errorf("Main: gitssh new public keys from file : %w", err)
			}
			c.GitAuth = authSSH
		}
	}
	return nil
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
		out.LatestOverallVersion = *v
		if current != nil && !v.GreaterThan(current) {
			continue
		}
		out.LatestOverallUpdate = *v
		if constraints == nil || !constraints.Check(v) {
			continue
		}
		out.LatestMatchingVersion = *v
		if current != nil {
			out.LatestMatchingUpdate = *v
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
		versions, err := versions.RegistryProvider(c.Registry, reg.Namespace, reg.Type)
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
