package transpiler

import (
	"encoding/xml"
	"fmt"
	"strings"

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

	if err := g.writeOutputDefinitions(program.Outputs); err != nil {
		return "", fmt.Errorf("error writing output definitions: %w", err)
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

	t.RegisterImplementationHandler("run_docker", t.handleDockerImplementation)
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

// These constants represent the types of outputs that can be generated in Galaxy.
// Data is for single files, DataCollection is for collections of files.
type GalaxyOutputType string

const (
	GalaxyOutputTypeData           GalaxyOutputType = "data"
	GalaxyOutputTypeDataCollection GalaxyOutputType = "collection"
)

// writeTypeValidation generates type validation code for parameters.
func (g *GalaxyTranspiler) writeTypeValidation(params []ast.Parameter) error {
	if len(params) == 0 {
		return nil
	}
	for _, param := range params {
		if param.Default != nil {
			continue
		}

		// Check for Galaxy Data Table metadata
		if tableName, ok := param.Metadata["galaxy_data_table"]; ok {
			if err := g.createDataTableParam(param, tableName); err != nil {
				return fmt.Errorf("error creating data table param '%s': %w", param.Name, err)
			}
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

func (g *GalaxyTranspiler) createDataTableParam(param ast.Parameter, tableName string) error {
	options := &galaxy.Options{
		FromDataTable: tableName,
		Columns: []galaxy.Column{
			{Name: "name", Index: 1},
			{Name: "value", Index: 0},
			{Name: "path", Index: 2},
		},
		Filter: []galaxy.Filter{
			{Type: "sort_by", Column: 1},
		},
	}

	g.galaxyTool.Inputs.Param = append(g.galaxyTool.Inputs.Param, galaxy.Param{
		Type:       "select",
		Name:       param.Name,
		Label:      param.Description,
		OptionsTag: options,
	})
	return nil
}

// writeOutputDefinitions generates output definitions for the Galaxy tool.
func (g *GalaxyTranspiler) writeOutputDefinitions(outputs []ast.OutputBlock) error {
	if len(outputs) == 0 {
		return nil
	}
	for _, output := range outputs {
		if output.Format == "directory" {
			g.galaxyTool.Outputs.Collection = append(g.galaxyTool.Outputs.Collection, galaxy.Collection{
				Name: output.Name,
				Type: "list", // Assuming "list" for now, as Baryon doesn't specify collection type
				Data: []galaxy.Data{
					{
						Name:   output.Name, // Use the output name for the data element inside the collection
						Format: "auto",      // Galaxy often uses 'auto' for collection elements
						Label:  output.Description,
					},
				},
			})
		} else {
			g.galaxyTool.Outputs.Data = append(g.galaxyTool.Outputs.Data, galaxy.Data{
				Name:   output.Name,
				Format: output.Format,
				Label:  output.Description,
			})
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

func (g *GalaxyTranspiler) handleDockerImplementation(
	t BaseTranspiler,
	impl *ast.ImplementationBlock,
	program *ast.Program) error {
	image, ok := impl.Fields["image"].(string)
	if !ok || image == "" {
		return fmt.Errorf("docker implementation requires 'image' option")
	}

	// Handle arguments
	args, ok := impl.Fields["arguments"].([]any)
	if ok && len(args) > 0 {
		for _, arg := range args {
			argStr, ok := arg.(string)
			if ok {
				// Format the argument to include Galaxy parameter references
				formattedArg := formatGalaxyArgument(argStr, program.Parameters)
				if g.galaxyTool.Command == nil {
					g.galaxyTool.Command = &galaxy.Command{
						Value: "",
					}
				}
				if g.galaxyTool.Command.Value != "" {
					g.galaxyTool.Command.Value += " "
				}
				g.galaxyTool.Command.Value += formattedArg
			}
		}
	}

	g.galaxyTool.Requirements.Container = []galaxy.Container{
		{
			Type:  "docker",
			Value: image,
		},
	}
	return nil
}

// formatGalaxyArgument checks if the given string is a Baryon parameter name
// and formats it into a Galaxy-compatible argument.
func formatGalaxyArgument(arg string, params []ast.Parameter) string {
	for _, param := range params {
		if param.Name == arg {
			// Check for Data Table metadata
			if _, ok := param.Metadata["galaxy_data_table"]; ok {
				return fmt.Sprintf("$%s.fields.path", param.Name)
			}

			// For file and directory types, Galaxy often uses .path or .name attributes
			// For simplicity, we start with $param_name. For directories, use .path
			if param.Type == TypeFile || param.Type == TypeDirectory {
				return fmt.Sprintf("$%s.path", param.Name)
			}
			return fmt.Sprintf("$%s", param.Name)
		}
	}
	// If it's not a parameter, and contains spaces, wrap in single quotes for basic shell safety
	if strings.ContainsAny(arg, " \t\n\r") {
		return fmt.Sprintf("'%s'", arg)
	}
	return arg
}
