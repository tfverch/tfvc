package checker

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"

	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/tfverch/tfvc/internal/lockfile"
	"github.com/tfverch/tfvc/internal/output"
	"github.com/tfverch/tfvc/internal/registry"
)

func Main(path string, includePrerelease bool, sshPrivKeyPath string, sshPrivKeyPwd string) (output.Updates, error) { //nolint:gocognit
	var updates output.Updates
	updatesClient := Client{
		Registry: registry.Client{
			HTTP: http.DefaultClient,
		},
	}
	mod, err := tfconfig.LoadModule(path)
	if err != nil {
		return nil, fmt.Errorf("reading root terraform module %q: %w", path, err)
	}
	locks := lockfile.LoadLocks(filepath.Join(path, ".terraform.lock.hcl"))
	if locks == nil {
		return nil, fmt.Errorf("Need to rewrite lockfile pkg error handling")
	}

	for _, provider := range mod.RequiredProviders {
		parsed, err := parseProvider(provider, locks)
		if err != nil {
			return nil, err
		}
		update, err := updatesClient.Update(*parsed.Source, parsed.Version, parsed.Constraints, includePrerelease)
		if err != nil {
			return nil, err
		}
		if update != nil { //nolint:all
		}
		updateOutput := output.Update{
			Type:               "provider",
			Path:               path,
			Name:               provider.Source,
			Source:             provider.Source,
			VersionConstraints: parsed.Constraints,
			LatestMatching:     update.LatestMatchingVersion,
			LatestOverall:      update.LatestOverallVersion,
		}
		if parsed.Version != nil {
			updateOutput.Version = *parsed.Version
		}
		updateOutput.SetUpdateStatus()
		updates = append(updates, updateOutput)
	}

	for _, module := range mod.ModuleCalls {
		parsed, err := parseModule(module)
		if err != nil {
			return nil, err
		}
		source := *parsed.Source
		if source.Local != nil {
			continue
		}
		if source.Git != nil {
			ssh, err := regexp.MatchString("git@", source.Git.Remote) //nolint:staticcheck
			if err != nil {
				return nil, err
			}
			if ssh {
				authSSH, err := gitssh.NewPublicKeysFromFile("git", sshPrivKeyPath, sshPrivKeyPwd)
				if err != nil {
					return nil, fmt.Errorf("CheckForUpdates: gitssh new public keys from file : %w", err)
				}
				updatesClient.GitAuth = authSSH
			}
		}

		update, err := updatesClient.Update(source, parsed.Version, parsed.Constraints, includePrerelease)
		if err != nil {
			return nil, err
		}
		updateOutput := output.Update{
			Type:               "module",
			Path:               module.Pos.Filename,
			Name:               module.Name,
			Source:             module.Source,
			VersionConstraints: parsed.Constraints,
			LatestMatching:     update.LatestMatchingVersion,
			LatestOverall:      update.LatestOverallVersion,
		}
		if parsed.Version != nil {
			updateOutput.Version = *parsed.Version
		}
		updateOutput.SetUpdateStatus()
		updates = append(updates, updateOutput)
	}
	return updates, nil
}
