package checker

import (
	"fmt"
	"net/http"
	"os"
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
	mod, diag := tfconfig.LoadModule(path)
	if diag.HasErrors() {
		return nil, fmt.Errorf("Main: reading root terraform module %q: %w", path, diag.Err())
	}
	lockfilepath := filepath.Join(path, ".terraform.lock.hcl")
	locks := &lockfile.Locks{}
	if _, err := os.Stat(lockfilepath); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("") // .terraform.lock.hcl not found. Need to actually write a warning function to advise to run terrafrom init.
		}
	} else {
		var loadErr error
		locks, loadErr = lockfile.LoadLocks(lockfilepath)
		if loadErr != nil {
			return nil, fmt.Errorf("Main: %w", err)
		}
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
				return nil, fmt.Errorf("Main: error checking git ssh regex %w", err)
			}
			if ssh {
				authSSH, err := gitssh.NewPublicKeysFromFile("git", sshPrivKeyPath, sshPrivKeyPwd)
				if err != nil {
					return nil, fmt.Errorf("Main: gitssh new public keys from file : %w", err)
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
