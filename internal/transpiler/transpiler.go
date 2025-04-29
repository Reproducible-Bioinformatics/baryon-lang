package transpiler

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/reproducible-bioinformatics/baryon-lang/internal/ast"
	"slices"
)

// Transpiler defines the interface for all language transpilers.
type Transpiler interface {
	// Transpile converts a Baryon program AST to target language code.
	Transpile(program *ast.Program) (string, error)
	// RegisterImplementationHandler adds a custom implementation handler.
	RegisterImplementationHandler(name string, handler ImplementationHandler)
	// RegisterTypeValidator adds a custom type validator.
	RegisterTypeValidator(typeName string, validator TypeValidator)
}

// ImplementationHandler processes implementation blocks.
type ImplementationHandler func(
	t BaseTranspiler, impl *ast.ImplementationBlock, program *ast.Program) error

// TypeValidator generates validation code for a parameter type.
type TypeValidator func(t BaseTranspiler, param ast.Parameter) error

// BaseTranspiler provides common functionality for all transpilers.
type BaseTranspiler interface {
	// Write a line of code with proper indentation.
	WriteLine(format string, args ...any)
	// Get the current indentation level.
	GetIndentLevel() int
	// Set the indentation level.
	SetIndentLevel(level int)
	// Get implementation handlers.
	GetImplementationHandlers() map[string]ImplementationHandler
	// Get type validators.
	GetTypeValidators() map[string]TypeValidator
	// Get the buffer containing the generated code.
	GetBuffer() *bytes.Buffer
}

// TranspilerBase implements BaseTranspiler and provides common functionality
type TranspilerBase struct {
	IndentLevel    int
	Buffer         bytes.Buffer
	ImplHandlers   map[string]ImplementationHandler
	TypeValidators map[string]TypeValidator
}

func (t *TranspilerBase) WriteLine(format string, args ...any) {
	indent := strings.Repeat("  ", t.IndentLevel)
	fmt.Fprintf(&t.Buffer, indent+format+"\n", args...)
}

func (t *TranspilerBase) GetIndentLevel() int {
	return t.IndentLevel
}

func (t *TranspilerBase) SetIndentLevel(level int) {
	t.IndentLevel = level
}

func (t *TranspilerBase) GetImplementationHandlers() map[string]ImplementationHandler {
	return t.ImplHandlers
}

func (t *TranspilerBase) GetTypeValidators() map[string]TypeValidator {
	return t.TypeValidators
}

func (t *TranspilerBase) GetBuffer() *bytes.Buffer {
	return &t.Buffer
}

// Initialize a transpiler base with common handlers and validators.
func (t *TranspilerBase) Initialize() {
	t.ImplHandlers = make(map[string]ImplementationHandler)
	t.TypeValidators = make(map[string]TypeValidator)
}

// RegisterImplementationHandler adds a custom implementation handler.
func (t *TranspilerBase) RegisterImplementationHandler(name string, handler ImplementationHandler) {
	t.ImplHandlers[name] = handler
}

// RegisterTypeValidator adds a custom type validator.
func (t *TranspilerBase) RegisterTypeValidator(typeName string, validator TypeValidator) {
	t.TypeValidators[typeName] = validator
}

// FormatDescription formats multi-line descriptions for documentation
func FormatDescription(desc string) string {
	lines := strings.Split(desc, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, " ")
}

// IdentifyFileParameters finds parameters that likely represent files or directories
func IdentifyFileParameters(params []ast.Parameter) []string {
	fileParams := []string{}

	for _, param := range params {
		// Check explicit type
		if param.Type == "file" || param.Type == "directory" {
			fileParams = append(fileParams, param.Name)
			continue
		}
	}

	return fileParams
}

// IsParamReference checks if a string is a parameter reference rather than a literal
func IsParamReference(s string, params []ast.Parameter) bool {
	for _, param := range params {
		if param.Name == s {
			return true
		}
	}
	return false
}

// GetParamType returns the type of a parameter by name
func GetParamType(name string, params []ast.Parameter) string {
	for _, param := range params {
		if param.Name == name {
			return param.Type
		}
	}
	return ""
}

// Contains checks if a string is in a slice
func Contains(slice []string, s string) bool {
	return slices.Contains(slice, s)
}
