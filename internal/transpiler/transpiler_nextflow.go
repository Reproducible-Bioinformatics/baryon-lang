package transpiler

import (
	"fmt"
	"strings"

	"github.com/reproducible-bioinformatics/baryon-lang/internal/ast"
)

func init() {
	RegisterTranspiler("nextflow", &TranspilerDescriptor{
		Extension:   ".nf",
		Display:     "NextFlow",
		Initializer: func() Transpiler { return NewNextflowTranspiler() },
	})
}

type NextflowTranspiler struct {
	TranspilerBase
}

func NewNextflowTranspiler() *NextflowTranspiler {
	t := &NextflowTranspiler{}
	t.Initialize()
	t.RegisterImplementationHandler("run_docker", t.handleDockerImplementation)
	return t
}

// Transpile converts a Baryon program AST to Nextflow DSL code.
func (n *NextflowTranspiler) Transpile(program *ast.Program) (string, error) {
	n.Buffer.Reset()

	// Write workflow header
	n.writeWorkflowHeader(program)

	// Write parameter declarations
	n.writeParameters(program.Parameters)

	// Write process blocks
	err := n.processImplementations(program)
	if err != nil {
		return "", fmt.Errorf("error processing implementations: %w", err)
	}

	// Write workflow definition
	n.writeWorkflow(program)

	return n.Buffer.String(), nil
}

func (n *NextflowTranspiler) writeWorkflowHeader(program *ast.Program) {
	n.WriteLine("// Nextflow Workflow: %s", program.Name)
	if program.Description != "" {
		desc := FormatDescription(program.Description)
		n.WriteLine("// %s", strings.ReplaceAll(desc, "\n", "\n// "))
	}
	n.WriteLine("")
}

func (n *NextflowTranspiler) writeParameters(params []ast.Parameter) {
	n.WriteLine("// Input Parameters")
	for _, param := range params {
		defaultStr := ""
		if param.Default != nil {
			defaultStr = fmt.Sprintf(" = %v", param.Default)
		}

		switch param.Type {
		case TypeString, TypeFile, TypeDirectory:
			n.WriteLine("params.%s = '%s'", param.Name, defaultStr)
		case TypeNumber, TypeInteger:
			n.WriteLine("params.%s = %s", param.Name, defaultStr)
		case TypeBoolean:
			n.WriteLine("params.%s = s", param.Name, defaultStr)
		case TypeEnum:
			if len(param.Constraints) > 0 {
				choices := make([]string, len(param.Constraints))
				for i, c := range param.Constraints {
					choices[i] = fmt.Sprintf("\"%v\"", c)
				}
				choicesStr := strings.Join(choices, ", ")
				n.WriteLine("// Allowed values: %s", choicesStr)
				n.WriteLine("params.%s = ''%s", param.Name, defaultStr)
			}
		case TypeCharacter:
			n.WriteLine("params.%s = ''%s", param.Name, defaultStr)
		}
	}
	n.WriteLine("")
}

func (n *NextflowTranspiler) processImplementations(program *ast.Program) error {
	if len(program.Implementations) == 0 {
		n.WriteLine("// No implementation blocks found")
		n.WriteLine("throw new Exception('No implementation defined for this workflow')")
		return nil
	}

	for _, impl := range program.Implementations {
		handler, ok := n.GetImplementationHandlers()[impl.Name]
		if !ok {
			return fmt.Errorf("no handler registered for implementation '%s'", impl.Name)
		}

		err := handler(n, &impl, program)
		if err != nil {
			return fmt.Errorf("error in implementation '%s': %w", impl.Name, err)
		}
	}

	return nil
}

func (n *NextflowTranspiler) handleDockerImplementation(t BaseTranspiler, impl *ast.ImplementationBlock, program *ast.Program) error {
	image, ok := impl.Fields["image"].(string)
	if !ok || image == "" {
		return fmt.Errorf("Docker image not specified or invalid")
	}

	n.WriteLine("")
	n.WriteLine("process %s {", impl.Name)
	n.SetIndentLevel(n.GetIndentLevel() + 1)
	n.WriteLine("container '%s'", image)

	// Declare input parameters
	n.WriteLine("input:")
	for _, param := range program.Parameters {
		n.WriteLine("val params.%s", param.Name)
	}

	// Declare output
	n.WriteLine("output:")
	n.WriteLine("path 'results/'")

	// Script block
	n.WriteLine("script:")
	n.SetIndentLevel(n.GetIndentLevel() + 1)
	n.WriteLine("def args = [")
	if args, ok := impl.Fields["arguments"].([]any); ok {
		for _, arg := range args {
			argStr := fmt.Sprintf("%v", arg)
			if IsParamReference(argStr, program.Parameters) {
				n.WriteLine("params.%s,", argStr)
			} else {
				n.WriteLine("'%s',", argStr)
			}
		}
	}
	n.WriteLine("].join(' ')")
	n.WriteLine("sh 'docker run --rm %s $args'", image)
	n.SetIndentLevel(n.GetIndentLevel() - 1)

	n.SetIndentLevel(n.GetIndentLevel() - 1)
	n.WriteLine("}")
	return nil
}

func (n *NextflowTranspiler) writeWorkflow(program *ast.Program) {
	n.WriteLine("workflow {")
	n.SetIndentLevel(n.GetIndentLevel() + 1)
	for _, impl := range program.Implementations {
		n.WriteLine("%s()", impl.Name)
	}
	n.SetIndentLevel(n.GetIndentLevel() - 1)
	n.WriteLine("}")
}
