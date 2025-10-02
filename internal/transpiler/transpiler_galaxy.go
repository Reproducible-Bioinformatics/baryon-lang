package transpiler

// TODO: MOVE TO THE NEW ARCHITECTURE

import (
	"github.com/reproducible-bioinformatics/baryon-lang/internal/ast"
	"github.com/reproducible-bioinformatics/baryon-lang/internal/galaxy"
)

func init() {
	RegisterTranspiler("galaxy", &TranspilerDescriptor{
		Extension:   ".xml",
		Display:     "Galaxy",
		Initializer: func() Transpiler { return NewGalaxyTranspiler() },
	})
}

// GalaxyTranspiler converts Baryon AST to Galaxy XML format.
type GalaxyTranspiler struct {
	TranspilerBase
	galaxyTool *galaxy.Tool
}

// RegisterImplementationHandler implements Transpiler.
// Subtle: this method shadows the method (TranspilerBase).RegisterImplementationHandler of GalaxyTranspiler.TranspilerBase.
func (g *GalaxyTranspiler) RegisterImplementationHandler(
	name string,
	handler ImplementationHandler,
) {
	g.ImplHandlers[name] = handler
}

// RegisterTypeValidator implements Transpiler.
// Subtle: this method shadows the method (TranspilerBase).RegisterTypeValidator of GalaxyTranspiler.TranspilerBase.
func (g *GalaxyTranspiler) RegisterTypeValidator(
	typeName string,
	validator TypeValidator,
) {
	g.TypeValidators[typeName] = validator
}

// Transpile implements Transpiler.
func (g *GalaxyTranspiler) Transpile(program *ast.Program) (string, error) {

	g.galaxyTool = &galaxy.Tool{
		Id:           program.Name,
		Name:         program.Name,
		Description:  program.Description,
		Requirements: &galaxy.Requirements{},
		Inputs: &galaxy.Inputs{
			Param: []galaxy.Param{},
		},
		Outputs: &galaxy.Outputs{},
	}

	if len(program.Implementations) == 0 {
		g.galaxyTool.Command = &galaxy.Command{Value: "echo 'No implementations provided'"}
	}

	for _, impl := range program.Implementations {
		if handler, ok := g.GetImplementationHandlers()[impl.Name]; ok {
			handler(g, &impl, program)
		}
	}

	return "", nil
}

// NewGalaxyTranspiler initializes a new Galaxy transpiler instance.
func NewGalaxyTranspiler() *GalaxyTranspiler {
	t := &GalaxyTranspiler{}
	t.Initialize()

	return t
}

func (g *GalaxyTranspiler) validateStringType(_ BaseTranspiler, param ast.Parameter) error {
	g.galaxyTool.Inputs.Param = append(g.galaxyTool.Inputs.Param, galaxy.Param{
		Type:            "text",
		Name:            param.Name,
		Value:           param.Default.(string),
		Label:           param.Description,
		RefreshOnChange: false,
	})
	return nil
}
