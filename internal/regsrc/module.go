package regsrc

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	svchost "github.com/hashicorp/terraform-svchost"
)

var (
	ErrInvalidModuleSource = errors.New("not a valid registry module source")

	// nameSubRe is the sub-expression that matches a valid module namespace or
	// name. It's strictly a super-set of what GitHub allows for user/org and
	// repo names respectively, but more restrictive than our original repo-name
	// regex which allowed periods but could cause ambiguity with hostname
	// prefixes. It does not anchor the start or end so it can be composed into
	// more complex RegExps below. Alphanumeric with - and _ allowed in non
	// leading or trailing positions. Max length 64 chars. (GitHub username is
	// 38 max.)
	nameSubRe = "[0-9A-Za-z](?:[0-9A-Za-z-_]{0,62}[0-9A-Za-z])?"

	// providerSubRe is the sub-expression that matches a valid provider. It
	// does not anchor the start or end so it can be composed into more complex
	// RegExps below. Only lowercase chars and digits are supported in practice.
	// Max length 64 chars.
	providerSubRe = "[0-9a-z]{1,64}"

	// moduleSourceRe is a regular expression that matches the basic
	// namespace/name/provider[//...] format for registry sources. It assumes
	// any FriendlyHost prefix has already been removed if present.
	moduleSourceRe = regexp.MustCompile(
		fmt.Sprintf("^(%s)\\/(%s)\\/(%s)(?:\\/\\/(.*))?$",
			nameSubRe, nameSubRe, providerSubRe))

	// NameRe is a regular expression defining the format allowed for namespace
	// or name fields in module registry implementations.
	NameRe = regexp.MustCompile("^" + nameSubRe + "$")

	// ProviderRe is a regular expression defining the format allowed for
	// provider fields in module registry implementations.
	ProviderRe = regexp.MustCompile("^" + providerSubRe + "$")

	// these hostnames are not allowed as registry sources, because they are
	// already special case module sources in terraform.
	disallowed = map[string]bool{
		"github.com":    true,
		"bitbucket.org": true,
	}
)

// Module describes a Terraform Registry Module source.
type Module struct {
	// RawHost is the friendly host prefix if one was present. It might be nil
	// if the original source had no host prefix which implies
	// PublicRegistryHost but is distinct from having an actual pointer to
	// PublicRegistryHost since it encodes the fact the original string didn't
	// include a host prefix at all which is significant for recovering actual
	// input not just normalized form. Most callers should access it with Host()
	// which will return public registry host instance if it's nil.
	RawHost      *FriendlyHost
	RawNamespace string
	RawName      string
	RawProvider  string
	RawSubmodule string
}

// NewModule construct a new module source from separate parts. Pass empty
// string if host or submodule are not needed.
func NewModule(host, namespace, name, provider, submodule string) (*Module, error) {
	m := &Module{
		RawNamespace: namespace,
		RawName:      name,
		RawProvider:  provider,
		RawSubmodule: submodule,
	}
	if host != "" {
		h := NewFriendlyHost(host)
		if h != nil {
			fmt.Println("HOST:", h)
			if !h.Valid() || disallowed[h.Display()] {
				return nil, ErrInvalidModuleSource
			}
		}
		m.RawHost = h
	}
	return m, nil
}

// ParseModuleSource attempts to parse source as a Terraform registry module
// source. If the string is not found to be in a valid format,
// ErrInvalidModuleSource is returned. Note that this can only be used on
// "input" strings, e.g. either ones supplied by the user or potentially
// normalised but in Display form (unicode). It will fail to parse a source with
// a punycoded domain since this is not permitted input from a user. If you have
// an already normalized string internally, you can compare it without parsing
// by comparing with the normalized version of the subject with the normal
// string equality operator.
func ParseModuleSource(source string) (*Module, error) {
	const fourMatch, fiveMatch = 4, 5
	// See if there is a friendly host prefix.
	host, rest := ParseFriendlyHost(source)
	if host != nil {
		if !host.Valid() || disallowed[host.Display()] {
			return nil, ErrInvalidModuleSource
		}
	}

	matches := moduleSourceRe.FindStringSubmatch(rest)
	if len(matches) < fourMatch {
		return nil, ErrInvalidModuleSource
	}

	m := &Module{
		RawHost:      host,
		RawNamespace: matches[1],
		RawName:      matches[2],
		RawProvider:  matches[3],
	}

	if len(matches) == fiveMatch {
		m.RawSubmodule = matches[4]
	}

	return m, nil
}

// Display returns the source formatted for display to the user in CLI or web
// output.
func (m *Module) Display() string {
	return m.formatWithPrefix(m.normalizedHostPrefix(m.Host().Display()), false)
}

// Normalized returns the source formatted for internal reference or comparison.
func (m *Module) Normalized() string {
	return m.formatWithPrefix(m.normalizedHostPrefix(m.Host().Normalized()), false)
}

// String returns the source formatted as the user originally typed it assuming
// it was parsed from user input.
func (m *Module) String() string {
	// Don't normalize public registry hostname - leave it exactly like the user
	// input it.
	hostPrefix := ""
	if m.RawHost != nil {
		hostPrefix = m.RawHost.String() + "/"
	}
	return m.formatWithPrefix(hostPrefix, true)
}

// Equal compares the module source against another instance taking
// normalization into account.
func (m *Module) Equal(other *Module) bool {
	return m.Normalized() == other.Normalized()
}

// Host returns the FriendlyHost object describing which registry this module is
// in. If the original source string had not host component this will return the
// PublicRegistryHost.
func (m *Module) Host() *FriendlyHost {
	if m.RawHost == nil {
		return PublicRegistryHost
	}
	return m.RawHost
}

func (m *Module) normalizedHostPrefix(host string) string {
	if m.Host().Equal(PublicRegistryHost) {
		return ""
	}
	return host + "/"
}

func (m *Module) formatWithPrefix(hostPrefix string, preserveCase bool) string {
	suffix := ""
	if m.RawSubmodule != "" {
		suffix = "//" + m.RawSubmodule
	}
	str := fmt.Sprintf("%s%s/%s/%s%s", hostPrefix, m.RawNamespace, m.RawName,
		m.RawProvider, suffix)

	// lower case by default
	if !preserveCase {
		return strings.ToLower(str)
	}
	return str
}

// Module returns just the registry ID of the module, without a hostname or
// suffix.
func (m *Module) Module() string {
	return fmt.Sprintf("%s/%s/%s", m.RawNamespace, m.RawName, m.RawProvider)
}

var ErrSvcHostComparison = errors.New("error in for comparison of svchost")

// SvcHost returns the svchost.Hostname for this module. Since FriendlyHost may
// contain an invalid hostname, this also returns an error indicating if it
// could be converted to a svchost.Hostname. If no host is specified, the
// default PublicRegistryHost is returned.
func (m *Module) SvcHost() (svchost.Hostname, error) {
	if m.RawHost == nil {
		comp, err := svchost.ForComparison(PublicRegistryHost.Raw)
		if err != nil {
			return "", fmt.Errorf("ErrSvcHostComparison %w : %s", ErrSvcHostComparison, err)
		}
		return comp, nil
	}
	comp, err := svchost.ForComparison(m.RawHost.Raw)
	if err != nil {
		return "", fmt.Errorf("ErrSvcHostComparison %w : %s", ErrSvcHostComparison, err)
	}
	return comp, nil
}
