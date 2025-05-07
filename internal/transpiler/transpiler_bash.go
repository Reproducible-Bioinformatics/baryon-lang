package transpiler

import "github.com/reproducible-bioinformatics/baryon-lang/internal/ast"

func init() {
	RegisterTranspiler("bash", &TranspilerDescriptor{
		Extension:   ".sh",
		Display:     "BASH",
		Initializer: func() Transpiler { return NewBashTranspiler() },
	})
}

// BashTranspiler converts Baryon AST to BASH code.
type BashTranspiler struct{ TranspilerBase }

// RegisterImplementationHandler implements Transpiler.
// Subtle: this method shadows the method
// (TranspilerBase).RegisterImplementationHandler of
// BashTranspiler.TranspilerBase.
func (b *BashTranspiler) RegisterImplementationHandler(
	name string,
	handler ImplementationHandler,
) {
	panic("unimplemented")
}

// RegisterTypeValidator implements Transpiler.
// Subtle: this method shadows the method
// (TranspilerBase).RegisterTypeValidator of BashTranspiler.TranspilerBase.
func (b *BashTranspiler) RegisterTypeValidator(
	typeName string,
	validator TypeValidator,
) {
	panic("unimplemented")
}

// Transpile implements Transpiler.
func (b *BashTranspiler) Transpile(program *ast.Program) (string, error) {
	panic("unimplemented")
}

func NewBashTranspiler() *BashTranspiler { return &BashTranspiler{} }
