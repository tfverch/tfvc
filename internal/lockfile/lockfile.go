package lockfile

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

func LoadLocks(path string) (*Locks, error) {
	locks := NewLocks()
	file, diag := hclparse.NewParser().ParseHCLFile(path)
	if diag.HasErrors() {
		return nil, fmt.Errorf("LoadLocks: %s : %w", diag.Error(), diag.Errs()[0])
	}
	var err error
	locks, err = loader(file, locks)
	if err != nil {
		return nil, fmt.Errorf("LoadLocks: %w", err)
	}
	return locks, nil
}

func loader(file *hcl.File, locks *Locks) (*Locks, error) {
	content, diag := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "provider",
				LabelNames: []string{"source_addr"},
			},
		},
	})
	if diag.HasErrors() {
		return nil, fmt.Errorf("loader: decode body content : %w", diag.Errs()[0])
	}

	seenProviders := make(map[Provider]hcl.Range)
	for _, block := range content.Blocks {
		switch block.Type {
		case "provider":
			lock, err := decodeProviderLockFromHCL(block)
			if err != nil {
				return nil, fmt.Errorf("loader: %w", err)
			}
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
	return locks, nil
}

func NewLocks() *Locks {
	return &Locks{
		Providers: make(map[Provider]*ProviderLock),
	}
}

func decodeProviderLockFromHCL(block *hcl.Block) (*ProviderLock, error) {
	ret := &ProviderLock{}
	rawAddr := block.Labels[0]
	addr, err := ParseProviderSource(rawAddr)
	if err != nil {
		return nil, fmt.Errorf("decodeProviderLockFromHCL: ParseProviderSource: %w", err)
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
		return nil, fmt.Errorf("decodeProviderLockFromHCL: body content to schema: %w", diags.Errs()[0])
	}

	v, verExists := content.Attributes["version"]
	if verExists {
		version, err := decodeProviderVersionArgument(v)
		if err == nil {
			ret.Version = version
		}
	}

	v, conExists := content.Attributes["constraints"]
	if conExists {
		constraints, err := decodeProviderVersionConstraintsArgument(v)
		if err == nil {
			ret.VersionConstraints = constraints
		}
	}

	return ret, nil
}

func decodeProviderVersionArgument(attr *hcl.Attribute) (Version, error) {
	expr := attr.Expr
	var raw *string
	hclDiags := gohcl.DecodeExpression(expr, nil, &raw)
	if hclDiags.HasErrors() {
		return Version{}, fmt.Errorf("decodeProviderVersionArgument: DecodeExpression: %w", hclDiags.Errs()[0])
	}
	version, err := ParseVersion(*raw)
	if err != nil {
		return Version{}, fmt.Errorf("decodeProviderVersionArgument: ParseVersion: %w", err)
	}
	return version, nil
}

func decodeProviderVersionConstraintsArgument(attr *hcl.Attribute) (VersionConstraints, error) {
	expr := attr.Expr
	var raw string
	hclDiags := gohcl.DecodeExpression(expr, nil, &raw)
	if hclDiags.HasErrors() {
		return nil, fmt.Errorf("decodeProviderVersionConstraintsArgument: DecodeExpression : %w", hclDiags.Errs()[0])
	}
	constraints, err := ParseVersionConstraints(raw)
	if err != nil {
		return nil, fmt.Errorf("decodeProviderVersionConstraintsArgument : ParseVersionConstraints : %w", err)
	}
	return constraints, nil
}
