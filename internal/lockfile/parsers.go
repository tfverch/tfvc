package lockfile

import (
	"fmt"
	"strings"

	svchost "github.com/hashicorp/terraform-svchost"
	"golang.org/x/net/idna"
)

const onePart, twoParts, threeParts = 1, 2, 3

func ParseProviderSource(str string) (Provider, error) {
	var ret Provider
	parts, err := parseSourceStringParts(str)
	if err != nil {
		return ret, err
	}

	name := parts[len(parts)-onePart]
	ret.Type = name
	ret.Hostname = DefaultProviderRegistryHost

	if len(parts) == onePart {
		return Provider{
			Hostname:  DefaultProviderRegistryHost,
			Namespace: UnknownProviderNamespace,
			Type:      name,
		}, nil
	}

	if len(parts) >= twoParts {
		// the namespace is always the second-to-last part
		givenNamespace := parts[len(parts)-twoParts]
		if givenNamespace == LegacyProviderNamespace {
			// For now we're tolerating legacy provider addresses until we've
			// finished updating the rest of the codebase to no longer use them,
			// or else we'd get errors round-tripping through legacy subsystems.
			ret.Namespace = LegacyProviderNamespace
		} else {
			namespace, err := ParseProviderPart(givenNamespace)
			if err != nil {
				return Provider{}, &ParserError{
					Summary: "Invalid provider namespace",
					Detail:  fmt.Sprintf(`Invalid provider namespace %q in source %q: %s"`, namespace, str, err),
				}
			}
			ret.Namespace = namespace
		}
	}

	// Final Case: 3 parts
	if len(parts) == threeParts {
		// the namespace is always the first part in a three-part source string
		hn, err := svchost.ForComparison(parts[0])
		if err != nil {
			return Provider{}, &ParserError{
				Summary: "Invalid provider source hostname",
				Detail:  fmt.Sprintf(`Invalid provider source hostname namespace %q in source %q: %s"`, hn, str, err),
			}
		}
		ret.Hostname = hn
	}

	if ret.Namespace == LegacyProviderNamespace && ret.Hostname != DefaultProviderRegistryHost {
		// Legacy provider addresses must always be on the default registry
		// host, because the default registry host decides what actual FQN
		// each one maps to.
		return Provider{}, &ParserError{
			Summary: "Invalid provider namespace",
			Detail:  "The legacy provider namespace \"-\" can be used only with hostname " + DefaultProviderRegistryHost.ForDisplay() + ".",
		}
	}

	// Due to how plugin executables are named and provider git repositories
	// are conventionally named, it's a reasonable and
	// apparently-somewhat-common user error to incorrectly use the
	// "terraform-provider-" prefix in a provider source address. There is
	// no good reason for a provider to have the prefix "terraform-" anyway,
	// so we've made that invalid from the start both so we can give feedback
	// to provider developers about the terraform- prefix being redundant
	// and give specialized feedback to folks who incorrectly use the full
	// terraform-provider- prefix to help them self-correct.
	const redundantPrefix = "terraform-"
	const userErrorPrefix = "terraform-provider-"
	if strings.HasPrefix(ret.Type, redundantPrefix) {
		if strings.HasPrefix(ret.Type, userErrorPrefix) {
			// Likely user error. We only return this specialized error if
			// whatever is after the prefix would otherwise be a
			// syntactically-valid provider type, so we don't end up advising
			// the user to try something that would be invalid for another
			// reason anyway.
			// (This is mainly just for robustness, because the validation
			// we already did above should've rejected most/all ways for
			// the suggestedType to end up invalid here.)
			suggestedType := ret.Type[len(userErrorPrefix):]
			if _, err := ParseProviderPart(suggestedType); err == nil {
				suggestedAddr := ret
				suggestedAddr.Type = suggestedType
				return Provider{}, &ParserError{
					Summary: "Invalid provider type",
					Detail:  fmt.Sprintf("Provider source %q has a type with the prefix %q, which isn't valid. Although that prefix is often used in the names of version control repositories for Terraform providers, provider source strings should not include it.\n\nDid you mean %q?", ret.ForDisplay(), userErrorPrefix, suggestedAddr.ForDisplay()),
				}
			}
		}
		// Otherwise, probably instead an incorrectly-named provider, perhaps
		// arising from a similar instinct to what causes there to be
		// thousands of Python packages on PyPI with "python-"-prefixed
		// names.
		return Provider{}, &ParserError{
			Summary: "Invalid provider type",
			Detail:  fmt.Sprintf("Provider source %q has a type with the prefix %q, which isn't allowed because it would be redundant to name a Terraform provider with that prefix. If you are the author of this provider, rename it to not include the prefix.", ret, redundantPrefix),
		}
	}

	return ret, nil
}

func parseSourceStringParts(str string) ([]string, error) {
	// split the source string into individual components
	parts := strings.Split(str, "/")
	if len(parts) == 0 || len(parts) > threeParts {
		return nil, &ParserError{
			Summary: "Invalid provider source string",
			Detail:  `The "source" attribute must be in the format "[hostname/][namespace/]name"`,
		}
	}

	// check for an invalid empty string in any part
	for i := range parts {
		if parts[i] == "" {
			return nil, &ParserError{
				Summary: "Invalid provider source string",
				Detail:  `The "source" attribute must be in the format "[hostname/][namespace/]name"`,
			}
		}
	}

	// check the 'name' portion, which is always the last part
	givenName := parts[len(parts)-1]
	name, err := ParseProviderPart(givenName)
	if err != nil {
		return nil, &ParserError{
			Summary: "Invalid provider type",
			Detail:  fmt.Sprintf(`Invalid provider type %q in source %q: %s"`, givenName, str, err),
		}
	}
	parts[len(parts)-onePart] = name

	return parts, nil
}

func ParseProviderPart(given string) (string, error) {
	if len(given) == 0 {
		return "", fmt.Errorf("ErrParseProviderPart %w : %s", ErrParseProviderPart, "must have at least one character")
	}

	// We're going to process the given name using the same "IDNA" library we
	// use for the hostname portion, since it already implements the case
	// folding rules we want.
	//
	// The idna library doesn't expose individual label parsing directly, but
	// once we've verified it doesn't contain any dots we can just treat it
	// like a top-level domain for this library's purposes.
	if strings.ContainsRune(given, '.') {
		return "", fmt.Errorf("ErrParseProviderPart %w : %s", ErrParseProviderPart, "dots are not allowed")
	}

	// We don't allow names containing multiple consecutive dashes, just as
	// a matter of preference: they look weird, confusing, or incorrect.
	// This also, as a side-effect, prevents the use of the "punycode"
	// indicator prefix "xn--" that would cause the IDNA library to interpret
	// the given name as punycode, because that would be weird and unexpected.
	if strings.Contains(given, "--") {
		return "", fmt.Errorf("ErrParseProviderPart %w : %s", ErrParseProviderPart, "cannot use multiple consecutive dashes")
	}

	result, err := idna.Lookup.ToUnicode(given)
	if err != nil {
		return "", fmt.Errorf("ErrParseProviderPart %w : %s", ErrParseProviderPart, "must contain only letters, digits, and dashes, and may not use leading or trailing dashes")
	}

	return result, nil
}
