package checker

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
		if parsed.RawCore != nil {
			output.Type = "terraform"
			output.Path = path
			output.Name = "terraform"
			output.Source = "github.com/hashicorp/terraform"
		}
		if parsed.RawProvider != nil {
			output.Type = "provider"
			output.Path = path
			output.Name = parsed.RawProvider.Source
			if parsed.Source.RegistryProvider != nil && parsed.Source.RegistryProvider.Normalized != "" {
				output.Source = parsed.Source.RegistryProvider.Normalized
			}
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
			return fmt.Errorf("checking git ssh regex %w", err)
		}
		if ssh {
			err := c.setSSHAuth(sshPrivKeyPath, sshPrivKeyPwd)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) setSSHAuth(sshPrivKeyPath string, sshPrivKeyPwd string) error {
	sshKeyPath, err := sshKeyPath(sshPrivKeyPath)
	if err != nil {
		return err
	}
	if sshKeyPath != "" {
		authSSH, err := gitssh.NewPublicKeysFromFile("git", sshKeyPath, sshPrivKeyPwd)
		if err != nil {
			return fmt.Errorf("Main: gitssh new public keys from file : %w", err)
		}
		c.GitAuth = authSSH
	}
	return nil
}

func sshKeyPath(sshPrivKeyPath string) (string, error) {
	if sshPrivKeyPath != "" {
		return sshPrivKeyPath, nil
	}
	names := []string{"id_rsa", "id_cdsa", "id_ecdsa_sk", "id_ed25519", "id_ed25519_sk", "id_dsa"}
	for _, name := range names {
		dirname, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path := filepath.Join(dirname, ".ssh", name)
		exists, err := exists(path)
		if err != nil {
			return "", err
		}
		if exists {
			return path, nil
		}
	}
	return "", nil
}

func exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("exists: checking file exists %w", err)
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
		versions = filterTerraformTags(s, versions)
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

func filterTerraformTags(s source.Source, versions []*goversion.Version) []*goversion.Version {
	// This is dumb but hashicorp have left two random tags in their git repo for v11 and 26258 lol!!
	// Here we simply remove these from the results.
	vers := []*goversion.Version{}
	if s.Git.Remote == "https://github.com/hashicorp/terraform.git" {
		for _, v := range versions {
			if v.String() != "11.0.0" && v.String() != "26258.0.0" {
				vers = append(vers, v)
			}
		}
		return vers
	}
	return versions
}
