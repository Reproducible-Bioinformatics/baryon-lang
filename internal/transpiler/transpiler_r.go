package transpiler

import (
	"fmt"
	"strings"

	"github.com/reproducible-bioinformatics/baryon-lang/internal/ast"
)

// RTranspiler converts Baryon AST to R code.
type RTranspiler struct {
	TranspilerBase
}

// NewRTranspiler creates a new RTranspiler instance with default handlers.
func NewRTranspiler() *RTranspiler {
	t := &RTranspiler{}
	t.Initialize()

	t.RegisterImplementationHandler("run_docker", t.handleDockerImplementation)

	typeValidators := map[string]TypeValidator{
		TypeString:    t.validateStringType,
		TypeNumber:    t.validateNumberType,
		TypeInteger:   t.validateIntegerType,
		TypeBoolean:   t.validateBooleanType,
		TypeEnum:      t.validateEnumType,
		TypeFile:      t.validateFileType,
		TypeDirectory: t.validateDirectoryType,
		TypeCharacter: t.validateCharacterType,
	}

	for name, fn := range typeValidators {
		t.RegisterTypeValidator(name, fn)
	}

	return t
}

// Transpile converts a Baryon program AST to R code
func (t *RTranspiler) Transpile(program *ast.Program) (string, error) {
	t.Buffer.Reset()

	t.writeDocumentation(program)

	t.writeSignature(program)

	err := t.writeTypeValidation(program.Parameters)
	if err != nil {
		return "", fmt.Errorf("error generating type validation: %w", err)
	}

	t.writeSecurityChecks(program.Parameters)

	err = t.processImplementations(program)
	if err != nil {
		return "", fmt.Errorf("error processing implementations: %w", err)
	}

	t.WriteLine("}")

	return t.Buffer.String(), nil
}

// writeDocumentation generates Roxygen-style documentation for the R function
func (t *RTranspiler) writeDocumentation(program *ast.Program) {
	t.WriteLine("#' %s", program.Name)
	t.WriteLine("#'")
	if program.Description != "" {
		t.WriteLine("#' @description %s", FormatDescription(program.Description))
	}

	// Parameter documentation
	for _, param := range program.Parameters {
		desc := param.Description
		if desc == "" {
			desc = fmt.Sprintf("Parameter of type '%s'", param.Type)
		}

		// For enum types, include allowed values
		if param.Type == "enum" && len(param.Constraints) > 0 {
			values := make([]string, len(param.Constraints))
			for i, v := range param.Constraints {
				values[i] = fmt.Sprintf("%v", v)
			}
			desc += fmt.Sprintf(" (allowed values: %s)", strings.Join(values, ", "))
		}

		t.WriteLine("#' @param %s %s", param.Name, FormatDescription(desc))
	}

	// Get return value from metadata if available
	returnDesc := "Results of the operation"
	if desc, ok := program.Metadata["return"]; ok {
		returnDesc = desc
	}
	t.WriteLine("#' @return %s", FormatDescription(returnDesc))
	t.WriteLine("#'")
	t.WriteLine("#' @export")
}

// writeSignature generates the function signature
func (t *RTranspiler) writeSignature(program *ast.Program) {
	// Create parameter list with default values where available
	params := make([]string, len(program.Parameters))
	for i, param := range program.Parameters {
		paramDef := param.Name
		if param.Default != nil {
			// Format default value based on type
			switch param.Type {
			case "string", "character":
				paramDef += fmt.Sprintf(" = \"%v\"", param.Default)
			case "boolean":
				boolVal, ok := param.Default.(bool)
				if ok {
					if boolVal {
						paramDef += " = TRUE"
					} else {
						paramDef += " = FALSE"
					}
				} else {
					paramDef += fmt.Sprintf(" = %v", param.Default)
				}
			default:
				paramDef += fmt.Sprintf(" = %v", param.Default)
			}
		}
		params[i] = paramDef
	}

	t.WriteLine("%s <- function(%s) {", program.Name, strings.Join(params, ",\n"))
	t.SetIndentLevel(t.GetIndentLevel() + 1)
}

// writeTypeValidation generates type checking code for all parameters
func (t *RTranspiler) writeTypeValidation(params []ast.Parameter) error {
	if len(params) == 0 {
		return nil
	}

	t.WriteLine("# Type validation")

	for _, param := range params {
		// Skip validation for parameters with default values
		if param.Default != nil {
			continue
		}

		validator, ok := t.GetTypeValidators()[param.Type]
		if !ok {
			// Default validation for unknown types
			t.WriteLine("# No specific validation for type '%s'", param.Type)
			continue
		}

		if err := validator(t, param); err != nil {
			return fmt.Errorf("error validating parameter '%s': %w", param.Name, err)
		}
	}

	return nil
}

// writeSecurityChecks generates security validation code
func (t *RTranspiler) writeSecurityChecks(params []ast.Parameter) {
	// Check for path traversal in file parameters
	fileParams := false
	for _, param := range params {
		if param.Type == "string" || param.Type == "file" || param.Type == "directory" {
			if !fileParams {
				t.WriteLine("")
				t.WriteLine("# Security checks")
				fileParams = true
			}

			t.WriteLine("if (grepl(\"\\\\.\\\\./|\\\\.\\\\\\\\|\\\\/\\\\.\\\\./|\\\\\\\\\\\\.\\\\\\\\\\\\.\\\\\\\\\", %s)) {", param.Name)
			t.SetIndentLevel(t.GetIndentLevel() + 1)
			t.WriteLine("stop(\"Path traversal detected in %s\")", param.Name)
			t.SetIndentLevel(t.GetIndentLevel() - 1)
			t.WriteLine("}")
		}
	}

	// Add file existence checks
	for _, param := range params {
		if param.Type == "file" {
			t.WriteLine("")
			t.WriteLine("# Check if file exists")
			t.WriteLine("if (!rrundocker::is_running_in_docker()) {")
			t.SetIndentLevel(t.GetIndentLevel() + 1)
			t.WriteLine("if (!file.exists(%s)) {", param.Name)
			t.SetIndentLevel(t.GetIndentLevel() + 1)
			t.WriteLine("stop(paste(\"%s:\", %s, \"does not exist\"))", param.Name, param.Name)
			t.SetIndentLevel(t.GetIndentLevel() - 1)
			t.WriteLine("}")
			t.SetIndentLevel(t.GetIndentLevel() - 1)
			t.WriteLine("}")
		} else if param.Type == "directory" {
			t.WriteLine("")
			t.WriteLine("# Check if directory exists")
			t.WriteLine("if (!rrundocker::is_running_in_docker()) {")
			t.SetIndentLevel(t.GetIndentLevel() + 1)
			t.WriteLine("if (!dir.exists(%s)) {", param.Name)
			t.SetIndentLevel(t.GetIndentLevel() + 1)
			t.WriteLine("stop(paste(\"%s:\", %s, \"does not exist\"))", param.Name, param.Name)
			t.SetIndentLevel(t.GetIndentLevel() - 1)
			t.WriteLine("}")
			t.SetIndentLevel(t.GetIndentLevel() - 1)
			t.WriteLine("}")
		}
	}
}

// processImplementations handles all implementation blocks
func (t *RTranspiler) processImplementations(program *ast.Program) error {
	if len(program.Implementations) == 0 {
		t.WriteLine("")
		t.WriteLine("# No implementation blocks found")
		t.WriteLine("stop(\"No implementation defined for this function\")")
		return nil
	}

	// Process each implementation block
	for _, impl := range program.Implementations {
		handler, ok := t.GetImplementationHandlers()[impl.Name]
		if !ok {
			return fmt.Errorf("no handler registered for implementation type '%s'", impl.Name)
		}

		if err := handler(t, &impl, program); err != nil {
			return fmt.Errorf("error processing '%s' implementation: %w", impl.Name, err)
		}
	}

	return nil
}

// validateStringType generates validation for string parameters
func (t *RTranspiler) validateStringType(base BaseTranspiler, param ast.Parameter) error {
	base.WriteLine("if (!is.character(%s) || length(%s) != 1) {", param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("stop(\"%s must be a single character string\")", param.Name)
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	base.WriteLine("}")

	return nil
}

// validateNumberType generates validation for numeric parameters
func (t *RTranspiler) validateNumberType(base BaseTranspiler, param ast.Parameter) error {
	base.WriteLine("if (!is.numeric(%s) || length(%s) != 1) {", param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("stop(\"%s must be a single numeric value\")", param.Name)
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	base.WriteLine("}")
	return nil
}

// validateIntegerType generates validation for integer parameters
func (t *RTranspiler) validateIntegerType(base BaseTranspiler, param ast.Parameter) error {
	base.WriteLine("if (!is.numeric(%s) || length(%s) != 1 || %s != round(%s)) {",
		param.Name, param.Name, param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("stop(\"%s must be a single integer value\")", param.Name)
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	base.WriteLine("}")
	return nil
}

// validateCharacterType generates validation for single character parameters
func (t *RTranspiler) validateCharacterType(base BaseTranspiler, param ast.Parameter) error {
	base.WriteLine("if (!is.character(%s) || length(%s) != 1 || nchar(%s) != 1) {",
		param.Name, param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("stop(\"%s must be a single character\")", param.Name)
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	base.WriteLine("}")
	return nil
}

// validateBooleanType generates validation for boolean parameters
func (t *RTranspiler) validateBooleanType(base BaseTranspiler, param ast.Parameter) error {
	base.WriteLine("if (!is.logical(%s) || length(%s) != 1) {", param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("stop(\"%s must be a single logical value (TRUE/FALSE)\")", param.Name)
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	base.WriteLine("}")
	return nil
}

// validateEnumType generates validation for enum parameters
func (t *RTranspiler) validateEnumType(base BaseTranspiler, param ast.Parameter) error {
	if len(param.Constraints) == 0 {
		return fmt.Errorf("enum type requires constraints with allowed values")
	}

	// Format constraint values
	constraints := make([]string, len(param.Constraints))
	for i, c := range param.Constraints {
		constraints[i] = fmt.Sprintf("\"%v\"", c)
	}

	// Generate validation code
	base.WriteLine("valid_%s <- c(%s)", param.Name, strings.Join(constraints, ", "))
	base.WriteLine("if (!is.character(%s) || length(%s) != 1 || !(%s %%in%% valid_%s)) {",
		param.Name, param.Name, param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("stop(paste0(\"%s must be one of: \", paste(valid_%s, collapse=\", \")))",
		param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	base.WriteLine("}")

	return nil
}

// validateFileType generates validation for file parameters
func (t *RTranspiler) validateFileType(base BaseTranspiler, param ast.Parameter) error {
	return t.validateStringType(base, param)

}

// validateDirectoryType generates validation for directory parameters
func (t *RTranspiler) validateDirectoryType(base BaseTranspiler, param ast.Parameter) error {
	return t.validateStringType(base, param)
}

// handleDockerImplementation generates code for Docker-based implementations
func (t *RTranspiler) handleDockerImplementation(base BaseTranspiler, impl *ast.ImplementationBlock, program *ast.Program) error {
	// Extract Docker configuration
	image, ok := impl.Fields["image"].(string)
	if !ok || image == "" {
		return fmt.Errorf("Docker image not specified or invalid")
	}

	base.WriteLine("")
	base.WriteLine("# Process file paths for Docker volume mounting")

	// Get file parameters for volume mounting
	fileParams := IdentifyFileParameters(program.Parameters)

	if len(fileParams) > 0 {
		// Setup for file parameters
		for _, param := range fileParams {
			base.WriteLine("# Process %s for Docker", param)
			base.WriteLine("%s_abspath <- normalizePath(%s, mustWork = FALSE)", param, param)
			base.WriteLine("%s_dir <- dirname(%s_abspath)", param, param)
			base.WriteLine("%s_filename <- basename(%s)", param, param)
		}

		// Use first file parameter's directory as main mount point
		base.WriteLine("")
		base.WriteLine("# Main volume mount point")
		base.WriteLine("main_mount_dir <- %s_dir", fileParams[0])
	} else {
		// Fallback to current directory
		base.WriteLine("# No file parameters found, using current directory")
		base.WriteLine("main_mount_dir <- normalizePath(getwd(), mustWork = FALSE)")
	}

	// Setup execution block with error handling
	base.WriteLine("")
	base.WriteLine("# Execute Docker container with error handling")
	base.WriteLine("tryCatch({")
	base.SetIndentLevel(base.GetIndentLevel() + 1)

	// Generate Docker run command
	base.WriteLine("result <- rrundocker::run_in_docker(")
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("image_name = \"%s\",", image)

	// Handle volumes
	volumes, ok := impl.Fields["volumes"].([]any)
	if ok && len(volumes) > 0 {
		base.WriteLine("volumes = list(")
		base.SetIndentLevel(base.GetIndentLevel() + 1)

		for _, vol := range volumes {
			switch v := vol.(type) {
			case []any:
				if len(v) >= 2 {
					// Handle volume specifications
					src := fmt.Sprintf("%v", v[0])
					dst := fmt.Sprintf("%v", v[1])

					// Check if src is a parameter reference
					if IsParamReference(src, program.Parameters) {
						base.WriteLine("c(%s_dir, \"%s\"),", src, dst)
					} else if src == "parent-folder" || src == "parent_folder" {
						base.WriteLine("c(main_mount_dir, \"%s\"),", dst)
					} else {
						base.WriteLine("c(\"%s\", \"%s\"),", src, dst)
					}
				}
			}
		}

		base.SetIndentLevel(base.GetIndentLevel() - 1)
		base.WriteLine("),")
	} else {
		// Default volume mapping if none specified
		base.WriteLine("volumes = list(")
		base.SetIndentLevel(base.GetIndentLevel() + 1)
		base.WriteLine("c(main_mount_dir, \"/data\")")
		base.SetIndentLevel(base.GetIndentLevel() - 1)
		base.WriteLine("),")
	}

	// Handle environment variables
	env, ok := impl.Fields["env"].([]any)
	if ok && len(env) > 0 {
		base.WriteLine("env = c(")
		base.SetIndentLevel(base.GetIndentLevel() + 1)

		for _, e := range env {
			switch ev := e.(type) {
			case []any:
				if len(ev) >= 2 {
					key := fmt.Sprintf("%v", ev[0])
					val := fmt.Sprintf("%v", ev[1])

					// Check if val is a parameter reference
					if IsParamReference(val, program.Parameters) {
						base.WriteLine("\"%s\" = %s,", key, val)
					} else {
						base.WriteLine("\"%s\" = \"%s\",", key, val)
					}
				}
			}
		}

		base.SetIndentLevel(base.GetIndentLevel() - 1)
		base.WriteLine("),")
	}

	// Handle arguments
	args, ok := impl.Fields["arguments"].([]any)
	if ok && len(args) > 0 {
		base.WriteLine("additional_arguments = c(")
		base.SetIndentLevel(base.GetIndentLevel() + 1)

		for _, arg := range args {
			argStr := fmt.Sprintf("%v", arg)

			// Skip placeholders
			if argStr == "_" {
				continue
			}

			// Check if it's a parameter reference
			if IsParamReference(argStr, program.Parameters) {
				paramType := GetParamType(argStr, program.Parameters)

				// Handle different parameter types
				if paramType == "file" || (paramType == "string" && Contains(fileParams, argStr)) {
					// Use just the filename for file parameters
					base.WriteLine("%s_filename,", argStr)
				} else if paramType == "number" || paramType == "integer" {
					// Convert numeric types to string
					base.WriteLine("as.character(%s),", argStr)
				} else if paramType == "boolean" {
					// Convert boolean to flag if TRUE
					base.WriteLine("if(%s) \"--true-flag\" else character(0),", argStr)
				} else {
					base.WriteLine("%s,", argStr)
				}
			} else if strings.HasPrefix(argStr, "\"") || strings.HasPrefix(argStr, "'") {
				// Already a string literal
				base.WriteLine("%s,", argStr)
			} else {
				// Treat as plain string
				base.WriteLine("\"%s\",", argStr)
			}
		}

		base.SetIndentLevel(base.GetIndentLevel() - 1)
		base.WriteLine(")")
	}

	base.SetIndentLevel(base.GetIndentLevel() - 1)
	base.WriteLine(")")

	// Process result
	base.WriteLine("")
	base.WriteLine("# Process result")
	base.WriteLine("return(list(")
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("status = \"success\",")
	base.WriteLine("output_dir = file.path(main_mount_dir, \"%s_results\")", program.Name)
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	base.WriteLine("))")

	// Error handling
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	base.WriteLine("}, error = function(e) {")
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("stop(paste(\"Docker execution failed:\", e$message))")
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	base.WriteLine("})")

	return nil
}
