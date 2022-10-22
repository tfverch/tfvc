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
				&source.Source{
					&source.Git{
						Remote: "https://github.com/hashicorp/terraform.git",
					},
					nil,
					nil,
					nil,
				},
				nil,
				"",
				cons,
				"~> 1.0",
				&RawCore{Name: "terraform", RequiredVersion: []string{"~> 1.0"}},
				nil,
				nil,
			},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := parseCore(test.raw)
			if err != nil {
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
