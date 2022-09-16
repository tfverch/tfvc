package checker

import (
	"fmt"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

type ProviderResult struct {
	ModulePath          string
	ProviderRequirement tfconfig.ProviderRequirement
}

type ModuleResult struct {
	Path       string
	ModuleCall tfconfig.ModuleCall
}

func scan(path string) ([]ProviderResult, []ModuleResult, error) {
	var providers []ProviderResult
	var modules []ModuleResult
	module, err := tfconfig.LoadModule(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read terraform module %q: %w", path, err)
	}
	for _, prov := range module.RequiredProviders {
		if prov == nil {
			continue
		}
		providers = append(providers, ProviderResult{
			ModulePath:          path,
			ProviderRequirement: *prov,
		})
	}
	for _, call := range module.ModuleCalls {
		if call == nil {
			continue
		}
		modules = append(modules, ModuleResult{
			Path:       call.Pos.Filename,
			ModuleCall: *call,
		})
	}
	return providers, modules, nil
}
