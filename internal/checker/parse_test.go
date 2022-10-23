package checker

import (
	"strings"
	"testing"

	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/stretchr/testify/assert"
	"github.com/tfverch/tfvc/internal/lockfile"
	"github.com/tfverch/tfvc/internal/source"
)

func TestParseCore(t *testing.T) {
	consSimple, _ := goversion.NewConstraint(strings.Join([]string{"~> 1.0"}, ","))
	tests := []struct {
		name        string
		input       []string
		expected    *Parsed
		shouldError bool
	}{
		{
			name:  "Simple ~> 1.0 constraint",
			input: []string{"~> 1.0"},
			expected: &Parsed{
				Source: &source.Source{
					Git: &source.Git{
						Remote: "https://github.com/hashicorp/terraform.git",
					},
				},
				Constraints:       consSimple,
				ConstraintsString: "~> 1.0",
				RawCore:           &RawCore{Name: "terraform", RequiredVersion: []string{"~> 1.0"}},
			},
			shouldError: false,
		},
		{
			name: "No constraint specified",
			expected: &Parsed{
				Source: &source.Source{
					Git: &source.Git{
						Remote: "https://github.com/hashicorp/terraform.git",
					},
				},
				RawCore: &RawCore{Name: "terraform"},
			},
			shouldError: false,
		},
		{
			name:        "Error malformed constraint string",
			input:       []string{"~>> 1.0"},
			expected:    nil,
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := parseCore(test.input)
			if err != nil {
				if test.shouldError {
					t.Logf("shouldError: %v", err)
				} else {
					t.Errorf("parseCore: %v", err)
					t.FailNow()
				}
			} else {
				assert.Equal(t, test.expected.Source, parsed.Source)
				assert.Equal(t, test.expected.Version, parsed.Version)
				assert.Equal(t, test.expected.VersionString, parsed.VersionString)
				assert.Equal(t, test.expected.Constraints.String(), parsed.Constraints.String())
				assert.Equal(t, test.expected.ConstraintsString, parsed.ConstraintsString)
				assert.Equal(t, test.expected.RawCore, parsed.RawCore)
				assert.Equal(t, test.expected.RawModule, parsed.RawModule)
				assert.Equal(t, test.expected.RawProvider, parsed.RawProvider)
			}
		})
	}
}

func TestParseProvider(t *testing.T) {
	type Input struct {
		key   string
		raw   *tfconfig.ProviderRequirement
		locks *lockfile.Locks
	}
	locks, _ := lockfile.LoadLocks("./test-data/.terraform.lock.hcl")
	cons, _ := goversion.NewConstraint(strings.Join([]string{"~> 4.0"}, ","))
	consPinned, _ := goversion.NewConstraint(strings.Join([]string{"4.0.0"}, ","))
	vers, _ := goversion.NewVersion("4.0.0")
	tests := []struct {
		name        string
		input       Input
		expected    *Parsed
		shouldError bool
	}{
		{
			name: "Simple ~> 4.0 constraint",
			input: Input{
				key: "registry.terraform.io/hashicorp/google",
				raw: &tfconfig.ProviderRequirement{
					Source:             "hashicorp/google",
					VersionConstraints: []string{"~> 4.0"},
				},
				locks: locks,
			},
			expected: &Parsed{
				Source: &source.Source{
					RegistryProvider: &source.RegistryProvider{
						Type:       "google",
						Namespace:  "hashicorp",
						Hostname:   svchost.Hostname("registry.terraform.io"),
						Normalized: "registry.terraform.io/hashicorp/google",
					},
				},
				Version:           vers,
				VersionString:     "4.0.0",
				Constraints:       cons,
				ConstraintsString: "~> 4.0",
				RawProvider: &tfconfig.ProviderRequirement{
					Source:             "hashicorp/google",
					VersionConstraints: []string{"~> 4.0"},
				},
			},
			shouldError: false,
		},
		{
			name: "Simple 4.0.0 constraint pinning to specific version",
			input: Input{
				key: "registry.terraform.io/hashicorp/google",
				raw: &tfconfig.ProviderRequirement{
					Source:             "hashicorp/google",
					VersionConstraints: []string{"4.0.0"},
				},
				locks: locks,
			},
			expected: &Parsed{
				Source: &source.Source{
					RegistryProvider: &source.RegistryProvider{
						Type:       "google",
						Namespace:  "hashicorp",
						Hostname:   svchost.Hostname("registry.terraform.io"),
						Normalized: "registry.terraform.io/hashicorp/google",
					},
				},
				Version:           vers,
				VersionString:     "4.0.0",
				Constraints:       consPinned,
				ConstraintsString: "4.0.0",
				RawProvider: &tfconfig.ProviderRequirement{
					Source:             "hashicorp/google",
					VersionConstraints: []string{"4.0.0"},
				},
			},
			shouldError: false,
		},
		{
			name: "No constraint specified",
			input: Input{
				key: "registry.terraform.io/hashicorp/google",
				raw: &tfconfig.ProviderRequirement{
					Source: "hashicorp/google",
				},
				locks: locks,
			},
			expected: &Parsed{
				Source: &source.Source{
					RegistryProvider: &source.RegistryProvider{
						Type:       "google",
						Namespace:  "hashicorp",
						Hostname:   svchost.Hostname("registry.terraform.io"),
						Normalized: "registry.terraform.io/hashicorp/google",
					},
				},
				RawProvider: &tfconfig.ProviderRequirement{
					Source: "hashicorp/google",
				},
			},
			shouldError: false,
		},
		{
			name: "No provider in lockfile",
			input: Input{
				key: "registry.terraform.io/hashicorp/aws",
				raw: &tfconfig.ProviderRequirement{
					Source: "hashicorp/aws",
				},
				locks: locks,
			},
			expected: &Parsed{
				Source: &source.Source{
					RegistryProvider: &source.RegistryProvider{
						Type:       "aws",
						Namespace:  "hashicorp",
						Hostname:   svchost.Hostname("registry.terraform.io"),
						Normalized: "registry.terraform.io/hashicorp/aws",
					},
				},
				RawProvider: &tfconfig.ProviderRequirement{
					Source: "hashicorp/aws",
				},
			},
			shouldError: false,
		},
		{
			name: "Empty provider source",
			input: Input{
				raw: &tfconfig.ProviderRequirement{
					Source: "",
				},
			},
			expected: &Parsed{
				Source: &source.Source{
					RegistryProvider: &source.RegistryProvider{},
				},
				RawProvider: &tfconfig.ProviderRequirement{},
			},
			shouldError: false,
		},
		{
			name: "Illegal provider source format",
			input: Input{
				raw: &tfconfig.ProviderRequirement{
					Source: "registry.terraform.io/hashicorp/aws/illegal",
				},
			},
			expected:    nil,
			shouldError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := parseProvider(test.input.key, test.input.raw, test.input.locks)
			if err != nil {
				if test.shouldError {
					t.Logf("shouldError: %v", err)
				} else {
					t.Errorf("Failed with the following unexpected error: %v", err)
					t.FailNow()
				}
			} else {
				assert.Equal(t, test.expected.Source, parsed.Source)
				assert.Equal(t, test.expected.Version, parsed.Version)
				assert.Equal(t, test.expected.VersionString, parsed.VersionString)
				assert.Equal(t, test.expected.Constraints.String(), parsed.Constraints.String())
				assert.Equal(t, test.expected.ConstraintsString, parsed.ConstraintsString)
				assert.Equal(t, test.expected.RawCore, parsed.RawCore)
				assert.Equal(t, test.expected.RawModule, parsed.RawModule)
				assert.Equal(t, test.expected.RawProvider, parsed.RawProvider)
			}
		})
	}
}

func TestParseModule(t *testing.T) {
	cons, _ := goversion.NewConstraint(strings.Join([]string{"v0.8.0"}, ","))
	vers, _ := goversion.NewVersion("v0.8.0")
	refValue := "v0.8.0"
	tests := []struct {
		name        string
		input       *tfconfig.ModuleCall
		expected    *Parsed
		shouldError bool
	}{
		{
			name: "Simple git module with ref",
			input: &tfconfig.ModuleCall{
				Source: "github.com/hashicorp/terraform-aws-consul?ref=v0.8.0",
			},
			expected: &Parsed{
				Source: &source.Source{
					Git: &source.Git{
						Remote:   "https://github.com/hashicorp/terraform-aws-consul.git",
						RefValue: &refValue,
					},
				},
				Version:       vers,
				VersionString: "v0.8.0",
				Constraints:   cons,
				RawModule: &tfconfig.ModuleCall{
					Source: "github.com/hashicorp/terraform-aws-consul?ref=v0.8.0",
				},
			},
			shouldError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := parseModule(test.input)
			if err != nil {
				if test.shouldError {
					t.Logf("shouldError: %v", err)
				} else {
					t.Errorf("Failed with the following unexpected error: %v", err)
					t.FailNow()
				}
			} else {
				assert.Equal(t, test.expected.Source, parsed.Source)
				assert.Equal(t, test.expected.Version, parsed.Version)
				assert.Equal(t, test.expected.VersionString, parsed.VersionString)
				assert.Equal(t, test.expected.Constraints.String(), parsed.Constraints.String())
				assert.Equal(t, test.expected.ConstraintsString, parsed.ConstraintsString)
				assert.Equal(t, test.expected.RawCore, parsed.RawCore)
				assert.Equal(t, test.expected.RawModule, parsed.RawModule)
				assert.Equal(t, test.expected.RawProvider, parsed.RawProvider)
			}
		})
	}
}
