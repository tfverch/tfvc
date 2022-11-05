package checker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/tfverch/tfvc/internal/lockfile"
	"github.com/tfverch/tfvc/internal/output"
)

var ErrNoTerraformModule = errors.New("no terraform module found")

func Main(path string, includePrerelease bool, sshPrivKeyPath string, sshPrivKeyPwd string) (output.Updates, error) {
	if !tfconfig.IsModuleDir(path) {
		return nil, fmt.Errorf("%w at %s", ErrNoTerraformModule, path)
	}
	mod, diag := tfconfig.LoadModule(path)
	if diag.HasErrors() {
		return nil, fmt.Errorf("main: reading root terraform module %q: %w", path, diag.Err())
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
	parsed, err := parse(mod, locks)
	if err != nil {
		return nil, err
	}
	updates, err := updates(parsed, includePrerelease, sshPrivKeyPath, sshPrivKeyPwd, path)
	if err != nil {
		return nil, err
	}
	return updates, nil
}
