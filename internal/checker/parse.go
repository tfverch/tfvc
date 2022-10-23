package checker

import (
	"fmt"
	"strings"

	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/tfverch/tfvc/internal/lockfile"
	"github.com/tfverch/tfvc/internal/source"
)

type Parsed struct {
	Source            *source.Source
	Version           *goversion.Version
	VersionString     string
	Constraints       goversion.Constraints
	ConstraintsString string
	RawCore           *RawCore
	RawModule         *tfconfig.ModuleCall
	RawProvider       *tfconfig.ProviderRequirement
}

type RawCore struct {
	Name            string
	RequiredVersion []string
}

func parse(root *tfconfig.Module, locks *lockfile.Locks) ([]Parsed, error) {
	parsedSlice := make([]Parsed, 0, len(root.RequiredProviders)+len(root.ModuleCalls)+1) // +1 for the terraform core version
	parsed, err := parseCore(root.RequiredCore)
	if err != nil {
		return nil, fmt.Errorf("parse required core: %w", err)
	}
	parsedSlice = append(parsedSlice, *parsed)
	for k, prov := range root.RequiredProviders {
		parsed, err := parseProvider(k, prov, locks)
		if err != nil {
			return nil, fmt.Errorf("parse: %w", err)
		}
		parsedSlice = append(parsedSlice, *parsed)
	}
	for _, mod := range root.ModuleCalls {
		parsed, err := parseModule(mod)
		if err != nil {
			return nil, fmt.Errorf("parse module call source: %w", err)
		}
		if parsed.Source.Local == nil {
			parsedSlice = append(parsedSlice, *parsed)
		}
	}
	return parsedSlice, nil
}

func parseCore(raw []string) (*Parsed, error) {
	src, err := source.Parse("github.com/hashicorp/terraform")
	if err != nil {
		return nil, fmt.Errorf("parse terraform core source: %w", err)
	}
	out := Parsed{Source: src, RawCore: &RawCore{Name: "terraform", RequiredVersion: raw}}
	if raw == nil {
		return &out, nil
	}
	constraintString := strings.Join(raw, ",")
	constraints, err := goversion.NewConstraint(constraintString)
	if err != nil {
		return nil, fmt.Errorf("parse constraint %q: %w", raw, err)
	}
	out.Constraints = constraints
	out.ConstraintsString = constraintString
	ver, err := goversion.NewVersion(constraintString)
	if err == nil { // interpret a single-version constraint as a pinned version
		out.Version = ver
		out.VersionString = raw[0]
	}
	return &out, nil
}

func parseProvider(key string, raw *tfconfig.ProviderRequirement, locks *lockfile.Locks) (*Parsed, error) {
	out := Parsed{Source: &source.Source{}, RawProvider: raw}
	src, err := source.ParseProviderSourceString(raw.Source)
	if err != nil {
		if raw.Source == "" {
			out.RawProvider.Source = key
			out.Source.RegistryProvider = &source.RegistryProvider{}
			return &out, nil
		}
		return nil, fmt.Errorf("parseProvider: %w", err)
	}
	out.Source = src
	if raw.VersionConstraints == nil {
		return &out, nil
	}
	constraintString := strings.Join(raw.VersionConstraints, ",")
	prov := lockfile.Provider{
		Namespace: src.RegistryProvider.Namespace,
		Type:      src.RegistryProvider.Type,
		Hostname:  src.RegistryProvider.Hostname,
	}
	if pr, ok := locks.Providers[prov]; ok {
		out.Version, err = goversion.NewVersion(pr.Version.String())
		if err != nil {
			return nil, fmt.Errorf("goversion new version %w", err)
		}
		out.VersionString = pr.Version.String()
	}
	ver, err := goversion.NewVersion(constraintString)
	if err == nil { // interpret a single-version constraint as a pinned version
		out.Version = ver
		out.VersionString = raw.VersionConstraints[0]
	}
	constraints, err := goversion.NewConstraint(constraintString)
	if err != nil {
		return nil, fmt.Errorf("parse constraint %q: %w", raw.VersionConstraints[0], err)
	}
	out.Constraints = constraints
	out.ConstraintsString = constraintString
	return &out, nil
}

func parseModule(raw *tfconfig.ModuleCall) (*Parsed, error) {
	src, err := source.Parse(raw.Source)
	if err != nil {
		return nil, fmt.Errorf("parse module call source: %w", err)
	}
	out := Parsed{Source: src, RawModule: raw}
	switch {
	case src.Git != nil:
		if ref := src.Git.RefValue; ref != nil {
			ver, err := goversion.NewVersion(*ref)
			if err == nil {
				out.Version = ver
			}
			out.VersionString = *ref
			constraints, err := goversion.NewConstraint(*ref)
			if err != nil {
				return nil, fmt.Errorf("parse constraint %q: %w", raw.Version, err)
			}
			out.Constraints = constraints
			out.ConstraintsString = raw.Version
		}
		if raw.Version == "" {
			return &out, nil
		}
	case src.Registry != nil:
		if raw.Version == "" {
			return &out, nil
		}
		ver, err := goversion.NewVersion(raw.Version)
		if err == nil { // interpret a single-version constraint as a pinned version
			out.Version = ver
			out.VersionString = raw.Version
		}
		constraints, err := goversion.NewConstraint(raw.Version)
		if err != nil {
			return nil, fmt.Errorf("parse constraint %q: %w", raw.Version, err)
		}
		out.Constraints = constraints
		out.ConstraintsString = raw.Version
	}
	return &out, nil
}
