package transpiler

import "github.com/reproducible-bioinformatics/baryon-lang/internal/ast"

func init() {
	RegisterTranspiler("galaxy", &TranspilerDescriptor{
		Extension:   ".xml",
		Display:     "Galaxy XML",
		Initializer: func() Transpiler { return NewGalaxyTranspiler() },
	})
}

// GalaxyTranspiler converts Baryon AST to Galaxy XML code.
type GalaxyTranspiler struct{ TranspilerBase }

// RegisterImplementationHandler implements Transpiler.
// Subtle: this method shadows the method
// (TranspilerBase).RegisterImplementationHandler of
// GalaxyTranspiler.TranspilerBase.
func (b *GalaxyTranspiler) RegisterImplementationHandler(
	name string,
	handler ImplementationHandler,
) {
	panic("unimplemented")
}

// RegisterTypeValidator implements Transpiler.
// Subtle: this method shadows the method
// (TranspilerBase).RegisterTypeValidator of GalaxyTranspiler.TranspilerBase.
func (b *GalaxyTranspiler) RegisterTypeValidator(
	typeName string,
	validator TypeValidator,
) {
	panic("unimplemented")
}

// Transpile implements Transpiler.
func (b *GalaxyTranspiler) Transpile(program *ast.Program) (string, error) {
	panic("unimplemented")
}

func NewGalaxyTranspiler() *GalaxyTranspiler { return &GalaxyTranspiler{} }
