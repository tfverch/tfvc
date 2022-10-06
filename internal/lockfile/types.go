package lockfile

import (
	"errors"
	"fmt"

	"github.com/apparentlymart/go-versions/versions"
	"github.com/apparentlymart/go-versions/versions/constraints"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/tfverch/tfvc/internal/regsrc"
)

type Provider = regsrc.Provider
type Version = versions.Version
type VersionConstraints = constraints.IntersectionSpec
type Locks struct {
	Providers           map[Provider]*ProviderLock
	OverriddenProviders map[Provider]struct{}
	Sources             map[string][]byte
}

type ProviderLock struct {
	// addr is the address of the provider this lock applies to.
	Addr               Provider
	Version            Version
	VersionConstraints VersionConstraints
}

type ParserError struct {
	Summary string
	Detail  string
}

func (pe *ParserError) Error() string {
	return fmt.Sprintf("%s: %s", pe.Summary, pe.Detail)
}

const DefaultProviderRegistryHost = svchost.Hostname("registry.terraform.io")
const UnknownProviderNamespace = "?"
const LegacyProviderNamespace = "-"

func ParseVersion(str string) (Version, error) {
	version, err := versions.ParseVersion(str)
	if err != nil {
		return version, fmt.Errorf("error parsing version %w", err)
	}
	return version, nil
}

func ParseVersionConstraints(str string) (VersionConstraints, error) {
	constraints, err := constraints.ParseRubyStyleMulti(str)
	if err != nil {
		return nil, fmt.Errorf("error parsing constraints %w", err)
	}
	return constraints, nil
}

var ErrParseProviderPart = errors.New("error parsing provider parts")
