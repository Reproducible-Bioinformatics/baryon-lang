package parser

import (
	"strings"
	"testing"

	"github.com/reproducible-bioinformatics/baryon-lang/internal/ast"
	"github.com/reproducible-bioinformatics/baryon-lang/internal/lexer"
)

func parseInput(input string) (*ast.Program, error) {
	l := lexer.New(input)
	p := New(l)
	return p.ParseProgram()
}

func TestParseProgram_ValidMinimal(t *testing.T) {
	input := `
	(bala myprog
		(
			(desc "A test program")
			(run_docker
				(image "ubuntu:latest")
				(command "echo hello")
			)
			(param1 string (desc "A string param"))
			(outputs
				(output.txt txt ./workdir/output.txt)
			)
		)
	)
	`
	prog, err := parseInput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prog.Name != "myprog" {
		t.Errorf("expected program name 'myprog', got %q", prog.Name)
	}
	if prog.Description != "A test program" {
		t.Errorf("expected description, got %q", prog.Description)
	}
	if len(prog.Implementations) != 1 {
		t.Errorf("expected 1 implementation, got %d", len(prog.Implementations))
	}
	if len(prog.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(prog.Parameters))
	}
	if len(prog.Outputs) != 1 {
		t.Errorf("expected 1 output, got %d", len(prog.Outputs))
	}
	t.Logf("Parsed program: %+v", prog)
}

func TestParseProgram_Invalid_NoBala(t *testing.T) {
	input := `
	(foo myprog
		(
			(desc "A test program")
		)
	)
	`
	_, err := parseInput(input)
	if err == nil || !strings.Contains(err.Error(), "program must start with 'bala'") {
		t.Errorf("expected error about 'bala', got %v", err)
	}
}

func TestParseParameterSExpr_Enum(t *testing.T) {
	input := `
	(bala myprog
		(
			(param1 (enum ("A" "B" "C")) (desc "enum param"))
		)
	)
	`
	prog, err := parseInput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(prog.Parameters))
	}
	param := prog.Parameters[0]
	if param.Type != "enum" {
		t.Errorf("expected type 'enum', got %q", param.Type)
	}
	if len(param.Constraints) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(param.Constraints))
	}
	if param.Description != "enum param" {
		t.Errorf("expected description, got %q", param.Description)
	}
}

func TestParseProgram_MissingParen(t *testing.T) {
	input := `
	(bala myprog
		(
			(desc "desc"
		)
	)
	`
	_, err := parseInput(input)
	if err == nil || !strings.Contains(err.Error(), "missing closing parenthesis") {
		t.Errorf("expected missing parenthesis error, got %v", err)
	}
}
