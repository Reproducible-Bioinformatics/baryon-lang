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
	for _, gt := range galaxyTypeValidators {
		t.RegisterTypeValidator(string(gt), t.validateGenericType(gt))
	}

	typeValidatorAlias := map[string]GalaxyTypeValidator{
		TypeString:    GalaxyTypeValidatorText,
		TypeCharacter: GalaxyTypeValidatorText,
		TypeNumber:    GalaxyTypeValidatorFloat,
		TypeInteger:   GalaxyTypeValidatorInteger,
		TypeBoolean:   GalaxyTypeValidatorBoolean,
		TypeFile:      GalaxyTypeValidatorFile,
		TypeDirectory: GalaxyTypeValidatorDataCollection,
	}

	for alias, gt := range typeValidatorAlias {
		t.RegisterTypeValidator(alias, t.validateGenericType(gt))
	}

	t.RegisterTypeValidator(TypeEnum, t.validateEnumType)

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

func (g *GalaxyTranspiler) validateGenericType(paramType GalaxyTypeValidator) func(BaseTranspiler, ast.Parameter) error {
	return func(_ BaseTranspiler, param ast.Parameter) error {
		g.galaxyTool.Inputs.Param = append(g.galaxyTool.Inputs.Param, galaxy.Param{
			Type:            string(paramType),
			Name:            param.Name,
			Value:           fmt.Sprintf("%v", param.Default),
			Label:           param.Description,
			RefreshOnChange: false,
		})
		return nil
	}
}

func (g *GalaxyTranspiler) validateEnumType(_ BaseTranspiler, param ast.Parameter) error {
	if len(param.Constraints) == 0 {
		return fmt.Errorf("enum type '%s' must have at least one constraint", param.Name)
	}

	opts := []galaxy.Option{}
	for _, opt := range param.Constraints {
		optString, ok := opt.(string)
		if !ok {
			continue
		}
		opts = append(opts, galaxy.Option{
			Value:         optString,
			CanonicalName: optString,
		})
	}

	g.galaxyTool.Inputs.Param = append(g.galaxyTool.Inputs.Param, galaxy.Param{
		Type:    string(GalaxyTypeValidatorSelect),
		Name:    param.Name,
		Label:   param.Description,
		Options: opts,
		Value:   opts[0].Value, // Default to first option
	})

	return nil
}
