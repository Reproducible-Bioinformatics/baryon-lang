package transpiler

import "github.com/reproducible-bioinformatics/baryon-lang/internal/ast"

func init() {
	RegisterTranspiler("streamflow", &TranspilerDescriptor{
		Extension:   "",
		Display:     "StreamFlow",
		Initializer: func() Transpiler { return NewStreamFlowTranspiler() },
	})
}

type StreamFlowTranspiler struct{ TranspilerBase }

// RegisterImplementationHandler implements Transpiler.
// Subtle: this method shadows the method (TranspilerBase).RegisterImplementationHandler of StreamFlowTranspiler.TranspilerBase.
func (s *StreamFlowTranspiler) RegisterImplementationHandler(name string, handler ImplementationHandler) {
	panic("unimplemented")
}

// RegisterTypeValidator implements Transpiler.
// Subtle: this method shadows the method (TranspilerBase).RegisterTypeValidator of StreamFlowTranspiler.TranspilerBase.
func (s *StreamFlowTranspiler) RegisterTypeValidator(typeName string, validator TypeValidator) {
	panic("unimplemented")
}

// Transpile implements Transpiler.
func (s *StreamFlowTranspiler) Transpile(program *ast.Program) (string, error) {
	panic("unimplemented")
}

func NewStreamFlowTranspiler() *StreamFlowTranspiler {

	return &StreamFlowTranspiler{}
}
