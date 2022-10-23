package checker

import (
	"testing"

	goversion "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/tfverch/tfvc/internal/source"
)

func TestParseCore(t *testing.T) {
	cons, _ := goversion.NewConstraint("~> 1.0")
	tests := []struct {
		name        string
		raw         []string
		expected    *Parsed
		shouldError bool
	}{
		{
			"Simple ~> 1.0 constraint",
			[]string{"~> 1.0"},
			&Parsed{
				Source: &source.Source{
					Git: &source.Git{
						Remote: "https://github.com/hashicorp/terraform.git",
					},
					Registry:         nil,
					RegistryProvider: nil,
					Local:            nil,
				},
				Version:           nil,
				VersionString:     "",
				Constraints:       cons,
				ConstraintsString: "~> 1.0",
				RawCore:           &RawCore{Name: "terraform", RequiredVersion: []string{"~> 1.0"}},
				RawModule:         nil,
				RawProvider:       nil,
			},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := parseCore(test.raw)
			if err != nil && !test.shouldError {
				t.Error("parseCore returned an unexpected error")
			}
			assert.Equal(t, test.expected.Source, parsed.Source)
			assert.Equal(t, test.expected.Version, parsed.Version)
			assert.Equal(t, test.expected.VersionString, parsed.VersionString)
			assert.Equal(t, test.expected.ConstraintsString, parsed.ConstraintsString)
			assert.Equal(t, test.expected.RawCore, parsed.RawCore)
			assert.Equal(t, test.expected.RawModule, parsed.RawModule)
			assert.Equal(t, test.expected.RawProvider, parsed.RawProvider)
		})
	}
}
