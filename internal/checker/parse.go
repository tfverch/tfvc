package checker

import (
	"fmt"
	"strings"

	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/ryan-jan/tfvc/internal/source"
)

type ParsedProvider struct {
	Source            *source.Source
	Version           *goversion.Version
	VersionString     string
	Constraints       goversion.Constraints
	ConstraintsString string
	Raw               tfconfig.ProviderRequirement
}

type ParsedModule struct {
	Source            *source.Source
	Version           *goversion.Version
	VersionString     string
	Constraints       goversion.Constraints
	ConstraintsString string
	Raw               tfconfig.ModuleCall
}

func parseProvider(raw tfconfig.ProviderRequirement) (*ParsedProvider, error) {
	parts := strings.Split(raw.Source, "/")
	src := &source.Source{
		RegistryProvider: &source.Registry{
			Namespace:  parts[0],
			Name:       parts[1],
			Normalized: raw.Source,
		},
	}
	out := ParsedProvider{Source: src, Raw: raw}
	if raw.VersionConstraints == nil {
		return &out, nil
	}
	constraintString := strings.Join(raw.VersionConstraints, ",")
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

func parseModule(raw tfconfig.ModuleCall) (*ParsedModule, error) {
	src, err := source.Parse(raw.Source)
	if err != nil {
		return nil, fmt.Errorf("parse module call source: %w", err)
	}
	out := ParsedModule{Source: src, Raw: raw}
	switch {
	case src.Git != nil:
		if ref := src.Git.RefValue; ref != nil {
			ver, err := goversion.NewVersion(*ref)
			if err == nil {
				out.Version = ver
			}
			out.VersionString = *ref
		}
		if raw.Version == "" {
			return &out, nil
		}
		// this adds (non-terraform-standard..) support for version constraints to Git sources
		constraints, err := goversion.NewConstraint(raw.Version)
		if err != nil {
			return nil, fmt.Errorf("parse constraint %q: %w", raw.Version, err)
		}
		out.Constraints = constraints
		out.ConstraintsString = raw.Version
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
