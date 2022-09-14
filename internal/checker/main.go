package checker

import (
	"log"
	"net/http"
	"regexp"

	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/ryan-jan/tfvc/internal/output"
	"github.com/ryan-jan/tfvc/internal/registry"
)

func CheckForUpdates(paths []string, includePrerelease bool, sshPrivKeyPath string, sshPrivKeyPwd string) []output.Update {
	var updates []output.Update
	updatesClient := Client{
		Registry: registry.Client{
			HTTP: http.DefaultClient,
		},
	}
	providerResults, moduleResults, err := scan(paths)
	if err != nil {
		log.Fatal(err)
	}

	for _, provider := range providerResults {
		parsed, err := parseProvider(provider.ProviderRequirement)
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}
		update, err := updatesClient.Update(*parsed.Source, parsed.Version, parsed.Constraints, includePrerelease)
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}
		updateOutput := output.Update{
			Path:              provider.ModulePath,
			Name:              provider.ProviderRequirement.Source,
			Source:            provider.ProviderRequirement.Source,
			VersionConstraint: parsed.ConstraintsString,
			Version:           parsed.VersionString,
			LatestMatching:    update.LatestMatchingVersion,
			MatchingUpdate:    update.LatestMatchingUpdate != "",
			LatestOverall:     update.LatestOverallVersion,
			NonMatchingUpdate: update.LatestOverallUpdate != "" && update.LatestOverallUpdate != update.LatestMatchingVersion,
		}
		updates = append(updates, updateOutput)
	}

	for _, module := range moduleResults {
		parsed, err := parseModule(module.ModuleCall)
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}
		source := *parsed.Source
		if source.Git != nil {
			ssh, err := regexp.MatchString("git@", source.Git.Remote)
			if err != nil {
				log.Fatal(err)
			}
			if ssh {
				authSsh, err := gitssh.NewPublicKeysFromFile("git", sshPrivKeyPath, sshPrivKeyPwd)
				if err != nil {
					log.Fatal(err)
				}
				updatesClient.GitAuth = authSsh
			}
		}

		update, err := updatesClient.Update(source, parsed.Version, parsed.Constraints, includePrerelease)
		if err != nil {
			log.Printf("error: %v", err)
			continue
		}

		updateOutput := output.Update{
			Path:              module.Path,
			Name:              module.ModuleCall.Name,
			Source:            module.ModuleCall.Source,
			VersionConstraint: parsed.ConstraintsString,
			Version:           parsed.VersionString,
			LatestMatching:    update.LatestMatchingVersion,
			MatchingUpdate:    update.LatestMatchingUpdate != "",
			LatestOverall:     update.LatestOverallVersion,
			NonMatchingUpdate: update.LatestOverallUpdate != "" && update.LatestOverallUpdate != update.LatestMatchingVersion,
		}
		updates = append(updates, updateOutput)
	}

	return updates
}
