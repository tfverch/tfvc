package lockfile

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

func LoadLocks(path string) *Locks {
	locks := NewLocks()
	file, diag := hclparse.NewParser().ParseHCLFile(path)
	if diag == nil {
		locks = loader(file, locks)
	}
	return locks
}

func loader(file *hcl.File, locks *Locks) *Locks {
	content, diag := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "provider",
				LabelNames: []string{"source_addr"},
			},
		},
	})
	if diag != nil {
		log.Fatal(diag)
	}

	seenProviders := make(map[Provider]hcl.Range)
	for _, block := range content.Blocks {

		switch block.Type {
		case "provider":
			lock := decodeProviderLockFromHCL(block)
			if lock == nil {
				continue
			}
			if previousRng, exists := seenProviders[lock.Addr]; exists {
				diag = diag.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate provider lock",
					Detail:   fmt.Sprintf("This lockfile already declared a lock for provider %s at %s.", lock.Addr.String(), previousRng.String()),
					Subject:  block.TypeRange.Ptr(),
				})
				continue
			}
			locks.Providers[lock.Addr] = lock
			seenProviders[lock.Addr] = block.DefRange

		default:
			// Shouldn't get here because this should be exhaustive for
			// all of the block types in the schema above.
		}
	}
	return locks
}

func NewLocks() *Locks {
	return &Locks{
		Providers: make(map[Provider]*ProviderLock),
	}
}

func decodeProviderLockFromHCL(block *hcl.Block) *ProviderLock {
	ret := &ProviderLock{}
	rawAddr := block.Labels[0]
	addr, err := ParseProviderSource(rawAddr)
	if err != nil {
		log.Fatal(err)
	}
	ret.Addr = addr

	content, diags := block.Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "version", Required: true},
			{Name: "constraints"},
			{Name: "hashes"},
		},
	})
	if diags.HasErrors() {
		log.Fatal(diags)
	}

	version := decodeProviderVersionArgument(addr, content.Attributes["version"])
	ret.Version = version

	constraints := decodeProviderVersionConstraintsArgument(addr, content.Attributes["constraints"])
	ret.VersionConstraints = constraints

	return ret
}

func decodeProviderVersionArgument(provider Provider, attr *hcl.Attribute) Version {
	expr := attr.Expr
	var raw *string
	hclDiags := gohcl.DecodeExpression(expr, nil, &raw)
	if hclDiags.HasErrors() {
		log.Fatal(hclDiags)
	}
	version, err := ParseVersion(*raw)
	if err != nil {
		log.Fatal(err)
	}
	return version
}

func decodeProviderVersionConstraintsArgument(provider Provider, attr *hcl.Attribute) VersionConstraints {
	expr := attr.Expr
	var raw string
	hclDiags := gohcl.DecodeExpression(expr, nil, &raw)
	if hclDiags.HasErrors() {
		log.Fatal(hclDiags)
	}
	constraints, err := ParseVersionConstraints(raw)
	if err != nil {
		log.Fatal(err)
	}
	return constraints
}
