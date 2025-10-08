package transpiler

import (
	"encoding/xml"
	"fmt"

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

	if err := g.writeTypeValidation(program.Parameters); err != nil {
		return "", fmt.Errorf("error writing type validation: %w", err)
	}

	if len(program.Implementations) == 0 {
		g.galaxyTool.Command = &galaxy.Command{
			Value: "echo 'No implementations provided'",
		}
	}

	for _, impl := range program.Implementations {
		if handler, ok := g.GetImplementationHandlers()[impl.Name]; ok {
			handler(g, &impl, program)
		}
	}

	outputString, err := xml.MarshalIndent(g.galaxyTool, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling Galaxy tool XML: %w", err)
	}
	return xml.Header + string(outputString), nil
}

// NewGalaxyTranspiler initializes a new Galaxy transpiler instance.
func NewGalaxyTranspiler() *GalaxyTranspiler {
	t := &GalaxyTranspiler{}
	t.Initialize()

	// t.RegisterImplementationHandler("run_docker", t.handleDockerImplementation)
	// t.RegisterImplementationHandler("run_singularity", t.handleSingularityImplementation)

	// Register type validators
	// https://docs.galaxyproject.org/en/latest/dev/schema.html#tool-inputs-param
	galaxyTypeValidators := []string{"text", "integer", "float", "boolean",
		"genomebuild", "select", "color", "data_column", "hidden",
		"hidden_data", "baseurl", "file", "ftpfile", "data", "data_collection",
		"drill_down"}
	for _, gt := range galaxyTypeValidators {
		t.RegisterTypeValidator(gt, t.validateGenericType(gt))
	}

	return t
}

// writeTypeValidation generates type validation code for parameters.
func (g *GalaxyTranspiler) writeTypeValidation(params []ast.Parameter) error {
	if len(params) == 0 {
		return nil
	}
	for _, param := range params {
		if param.Default != nil {
			continue
		}
		validator, exists := g.GetTypeValidators()[param.Type]
		if !exists {
			return fmt.Errorf("no validator registered for type '%s'", param.Type)
		}
		if err := validator(g, param); err != nil {
			return fmt.Errorf("error validating parameter '%s': %w", param.Name, err)
		}
	}
	return nil
}

func (g *GalaxyTranspiler) validateGenericType(paramType string) func(BaseTranspiler, ast.Parameter) error {
	return func(_ BaseTranspiler, param ast.Parameter) error {
		g.galaxyTool.Inputs.Param = append(g.galaxyTool.Inputs.Param, galaxy.Param{
			Type:            paramType,
			Name:            param.Name,
			Value:           fmt.Sprintf("%v", param.Default),
			Label:           param.Description,
			RefreshOnChange: false,
		})
		return nil
	}
}
