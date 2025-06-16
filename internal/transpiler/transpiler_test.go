package transpiler

import (
	"testing"

	"github.com/reproducible-bioinformatics/baryon-lang/internal/ast"
)

func TestFormatDescription(t *testing.T) {
	input := "This is a description.\nWith multiple lines.\n  And extra spaces. "
	expected := "This is a description. With multiple lines. And extra spaces."
	got := FormatDescription(input)
	if got != expected {
		t.Errorf("FormatDescription() = %q, want %q", got, expected)
	}
}

func TestIdentifyFileParameters(t *testing.T) {
	params := []ast.Parameter{
		{NamedBaseNode: ast.NamedBaseNode{Name: "input"}, Type: "file"},
		{NamedBaseNode: ast.NamedBaseNode{Name: "output"}, Type: "directory"},
		{NamedBaseNode: ast.NamedBaseNode{Name: "count"}, Type: "integer"},
	}
	got := IdentifyFileParameters(params)
	expected := []string{"input", "output"}
	if len(got) != len(expected) {
		t.Fatalf("IdentifyFileParameters() = %v, want %v", got, expected)
	}
	for i := range got {
		if got[i] != expected[i] {
			t.Errorf("IdentifyFileParameters()[%d] = %q, want %q", i, got[i], expected[i])
		}
	}
}

func TestIsParamReference(t *testing.T) {
	params := []ast.Parameter{
		{NamedBaseNode: ast.NamedBaseNode{Name: "foo"}, Type: "string"},
		{NamedBaseNode: ast.NamedBaseNode{Name: "bar"}, Type: "number"},
	}
	tests := []struct {
		input    string
		expected bool
	}{
		{"foo", true},
		{"bar", true},
		{"baz", false},
	}
	for _, tt := range tests {
		got := IsParamReference(tt.input, params)
		if got != tt.expected {
			t.Errorf("IsParamReference(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestGetParamType(t *testing.T) {
	params := []ast.Parameter{
		{NamedBaseNode: ast.NamedBaseNode{Name: "x"}, Type: "integer"},
		{NamedBaseNode: ast.NamedBaseNode{Name: "y"}, Type: "string"},
	}
	if typ := GetParamType("x", params); typ != "integer" {
		t.Errorf("GetParamType(x) = %q, want %q", typ, "integer")
	}
	if typ := GetParamType("y", params); typ != "string" {
		t.Errorf("GetParamType(y) = %q, want %q", typ, "string")
	}
	if typ := GetParamType("z", params); typ != "" {
		t.Errorf("GetParamType(z) = %q, want empty string", typ)
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	if !Contains(slice, "a") {
		t.Error("Contains(slice, \"a\") = false, want true")
	}
	if Contains(slice, "z") {
		t.Error("Contains(slice, \"z\") = true, want false")
	}
}

func TestTranspilerBase_WriteLine(t *testing.T) {
	tb := &TranspilerBase{}
	tb.SetIndentLevel(2)
	tb.WriteLine("hello %s", "world")
	expected := "    hello world\n"
	if tb.Buffer.String() != expected {
		t.Errorf("WriteLine() = %q, want %q", tb.Buffer.String(), expected)
	}
}

func TestTranspilerBase_IndentLevel(t *testing.T) {
	tb := &TranspilerBase{}
	tb.SetIndentLevel(3)
	if tb.GetIndentLevel() != 3 {
		t.Errorf("GetIndentLevel() = %d, want 3", tb.GetIndentLevel())
	}
}

func TestTranspilerBase_HandlersAndValidators(t *testing.T) {
	tb := &TranspilerBase{}
	tb.Initialize()
	handler := func(t BaseTranspiler, impl *ast.ImplementationBlock, program *ast.Program) error { return nil }
	validator := func(t BaseTranspiler, param ast.Parameter) error { return nil }
	tb.RegisterImplementationHandler("test", handler)
	tb.RegisterTypeValidator("foo", validator)
	if tb.GetImplementationHandlers()["test"] == nil {
		t.Error("Handler not registered")
	}
	if tb.GetTypeValidators()["foo"] == nil {
		t.Error("Validator not registered")
	}
}

func TestTranspilerBase_GetBuffer(t *testing.T) {
	tb := &TranspilerBase{}
	tb.Buffer.WriteString("abc")
	buf := tb.GetBuffer()
	if buf.String() != "abc" {
		t.Errorf("GetBuffer() = %q, want %q", buf.String(), "abc")
	}
}
