package transpiler

import (
	"strings"
	"testing"

	"github.com/reproducible-bioinformatics/baryon-lang/internal/ast"
)

func TestGalaxyDataTableParam(t *testing.T) {
	// Setup transpiler
	tr, err := GetTranspiler("galaxy")
	if err != nil {
		t.Fatalf("Failed to get galaxy transpiler: %v", err)
	}
	transpiler := tr.Initializer()

	// Create AST
	prog := &ast.Program{
		NamedBaseNode: ast.NamedBaseNode{Name: "test_tool"},
		Parameters: []ast.Parameter{
			{
				NamedBaseNode: ast.NamedBaseNode{
					Name: "ref_genome",
					BaseNode: ast.BaseNode{
						Description: "Reference Genome",
					},
				},
				Type: "file",
				Metadata: map[string]string{
					"galaxy_data_table": "fasta_indexes",
				},
			},
		},
		Implementations: []ast.ImplementationBlock{
            {
                BaseNode: ast.BaseNode{},
                Name:     "run_docker",
                Fields: map[string]any{
                    "image": "ubuntu",
                    "arguments": []any{"ref_genome"},
                },
            },
        },
	}

	// Transpile
	output, err := transpiler.Transpile(prog)
	if err != nil {
		t.Fatalf("Transpile failed: %v", err)
	}

	// Verify Output
	if !strings.Contains(output, `<options from_data_table="fasta_indexes">`) {
		t.Error("Output missing options tag with from_data_table")
	}
	if !strings.Contains(output, `<column name="path" index="2"></column>`) {
		t.Error("Output missing column definition")
	}
    // Check for the command string containing the correct variable expansion
    // The command is wrapped in <command> tags, often with specific quoting or layout
    if !strings.Contains(output, `$ref_genome.fields.path`) {
        t.Errorf("Output missing formatted argument. Got: %s", output)
    }
}
