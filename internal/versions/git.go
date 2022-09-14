package versions

import (
	"fmt"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	goversion "github.com/hashicorp/go-version"
)

func Git(remoteURL string, auth transport.AuthMethod) ([]*goversion.Version, error) {
	raw, err := git.Init(memory.NewStorage(), nil)
	if err != nil {
		return nil, fmt.Errorf("git init: %w", err)
	}
	remote, err := raw.CreateRemote(&config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{remoteURL},
	})
	if err != nil {
		return nil, fmt.Errorf("git remote: %w", err)
	}
	refs, err := remote.List(&git.ListOptions{Auth: auth})
	if err != nil {
		return nil, fmt.Errorf("git list refs: %w", err)
	}
	out := make([]*goversion.Version, 0, len(refs))
	for _, ref := range refs {
		version, err := goversion.NewVersion(ref.Name().Short())
		if err != nil {
			continue
		}
		out = append(out, version)
	}
	sort.Sort(goversion.Collection(out))
	return out, nil
}
