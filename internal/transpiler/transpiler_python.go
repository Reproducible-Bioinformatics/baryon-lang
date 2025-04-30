package transpiler

import (
	"fmt"
	"strings"

	"github.com/reproducible-bioinformatics/baryon-lang/internal/ast"
)

// PythonTranspiler converts Baryon's ast.Program to Python code.
type PythonTranspiler struct {
	TranspilerBase
}

// NewPythonTranspiler creates a new PythonTranspiler instance with default handlers.
func NewPythonTranspiler() *PythonTranspiler {
	t := &PythonTranspiler{}
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

// Transpile converts a Baryon program AST to Python code
func (t *PythonTranspiler) Transpile(program *ast.Program) (string, error) {
	t.Buffer.Reset()

	// Generate shebang and imports
	t.writeHeader()

	// Generate utility functions
	t.writeUtilityFunctions()

	// Generate function with docstring
	t.writeFunctionHeader(program)

	// Generate parameter validation
	err := t.writeTypeValidation(program.Parameters)
	if err != nil {
		return "", fmt.Errorf("error generating type validation: %w", err)
	}

	// Generate security checks
	t.writeSecurityChecks(program.Parameters)

	// Process implementation blocks
	err = t.processImplementations(program)
	if err != nil {
		return "", fmt.Errorf("error processing implementations: %w", err)
	}

	// Add main entry point
	t.SetIndentLevel(0)
	t.writeEntryPoint(program)

	return t.Buffer.String(), nil
}

// writeHeader generates header comments, shebang, and imports
func (t *PythonTranspiler) writeHeader() {
	t.WriteLine("#!/usr/bin/env python3")
	t.WriteLine("")
	t.WriteLine("import os")
	t.WriteLine("import sys")
	t.WriteLine("import re")
	t.WriteLine("import subprocess")
	t.WriteLine("import pathlib")
	t.WriteLine("import logging")
	t.WriteLine("from typing import Dict, List, Any, Optional, Union")
	t.WriteLine("from dataclasses import dataclass")
	t.WriteLine("")
	t.WriteLine("# Configure logging")
	t.WriteLine("logger = logging.getLogger(__name__)")
	t.WriteLine("")
}

// writeUtilityFunctions generates helper functions
func (t *PythonTranspiler) writeUtilityFunctions() {
	// Result dataclass
	t.WriteLine("@dataclass")
	t.WriteLine("class Result:")
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	t.WriteLine("status: str")
	t.WriteLine("output_dir: str")
	t.WriteLine("message: str = \"\"")
	t.SetIndentLevel(t.GetIndentLevel() - 1)
	t.WriteLine("")

	// Path validation function
	t.WriteLine("def validate_path(path: str) -> str:")
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	t.WriteLine("\"\"\"Validate and normalize a file path.\"\"\"")
	t.WriteLine("if not path:")
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	t.WriteLine("raise ValueError(\"Path cannot be empty\")")
	t.SetIndentLevel(t.GetIndentLevel() - 1)
	t.WriteLine("return os.path.abspath(os.path.expanduser(path))")
	t.SetIndentLevel(t.GetIndentLevel() - 1)
	t.WriteLine("")

	// Docker check function
	t.WriteLine("def is_running_in_docker() -> bool:")
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	t.WriteLine("\"\"\"Check if we're running inside a Docker container.\"\"\"")
	t.WriteLine("return os.path.exists('/.dockerenv')")
	t.SetIndentLevel(t.GetIndentLevel() - 1)
	t.WriteLine("")

	// Docker run function
	t.WriteLine("def run_docker(image: str, volumes: Dict[str, str], env: Dict[str, str], args: List[str]) -> str:")
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	t.WriteLine("\"\"\"Run a Docker container with specified parameters.\"\"\"")
	t.WriteLine("cmd = ['docker', 'run', '--rm']")
	t.WriteLine("")
	t.WriteLine("for src, dst in volumes.items():")
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	t.WriteLine("cmd.extend(['-v', f\"{src}:{dst}\"])")
	t.SetIndentLevel(t.GetIndentLevel() - 1)

	t.WriteLine("")
	t.WriteLine("for key, val in env.items():")
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	t.WriteLine("cmd.extend(['-e', f\"{key}={val}\"])")
	t.SetIndentLevel(t.GetIndentLevel() - 1)

	t.WriteLine("")
	t.WriteLine("cmd.append(image)")
	t.WriteLine("cmd.extend(args)")

	t.WriteLine("")
	t.WriteLine("logger.info(f\"Running Docker command: {' '.join(cmd)}\")")
	t.WriteLine("result = subprocess.run(cmd, capture_output=True, text=True, check=False)")

	t.WriteLine("")
	t.WriteLine("if result.returncode != 0:")
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	t.WriteLine("logger.error(f\"Docker execution failed: {result.stderr}\")")
	t.WriteLine("raise RuntimeError(f\"Docker execution failed: {result.stderr}\")")
	t.SetIndentLevel(t.GetIndentLevel() - 1)

	t.WriteLine("")
	t.WriteLine("return result.stdout")
	t.SetIndentLevel(t.GetIndentLevel() - 1)
	t.WriteLine("")
}

// writeFunctionHeader generates the function signature and docstring
func (t *PythonTranspiler) writeFunctionHeader(program *ast.Program) {
	// Generate function signature
	paramList := t.formatParameterList(program.Parameters)
	t.WriteLine("def %s(%s) -> Result:", program.Name, paramList)

	// Generate function docstring
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	t.WriteLine("\"\"\"")
	if program.Description != "" {
		t.WriteLine("%s", FormatDescription(program.Description))
		t.WriteLine("")
	}

	// Parameter documentation
	if len(program.Parameters) > 0 {
		t.WriteLine("Parameters:")
		for _, param := range program.Parameters {
			desc := param.Description
			if desc == "" {
				desc = fmt.Sprintf("Parameter of type '%s'", param.Type)
			}

			// For enum types, include allowed values
			if param.Type == "enum" && len(param.Constraints) > 0 {
				values := make([]string, len(param.Constraints))
				for i, c := range param.Constraints {
					values[i] = fmt.Sprintf("%v", c)
				}
				desc += fmt.Sprintf(" (allowed values: %s)", strings.Join(values, ", "))
			}

			t.WriteLine("    %s: %s", param.Name, FormatDescription(desc))
		}
		t.WriteLine("")
	}

	// Return value documentation
	returnDesc := "Results of the operation"
	if desc, ok := program.Metadata["return"]; ok {
		returnDesc = desc
	}
	t.WriteLine("Returns:")
	t.WriteLine("    Result: %s", FormatDescription(returnDesc))
	t.WriteLine("\"\"\"")
}

// formatParameterList generates a Python parameter list with type annotations
func (t *PythonTranspiler) formatParameterList(params []ast.Parameter) string {
	if len(params) == 0 {
		return ""
	}

	paramStrings := make([]string, len(params))
	for i, param := range params {
		// Build parameter with type annotation
		paramStr := param.Name + ": "

		// Add appropriate type based on Baryon type
		switch param.Type {
		case "string":
			paramStr += "str"
		case "number":
			paramStr += "float"
		case "integer":
			paramStr += "int"
		case "boolean":
			paramStr += "bool"
		case "file", "directory":
			paramStr += "str" // File paths are strings
		case "character":
			paramStr += "str" // Single character as string
		case "enum":
			paramStr += "str" // Enum as string with specific values
		default:
			paramStr += "Any"
		}

		// Add default value if specified
		if param.Default != nil {
			switch param.Type {
			case "string", "file", "directory", "character", "enum":
				paramStr += fmt.Sprintf(" = \"%v\"", param.Default)
			case "boolean":
				boolVal, ok := param.Default.(bool)
				if ok {
					paramStr += fmt.Sprintf(" = %v", boolVal)
				} else {
					paramStr += fmt.Sprintf(" = %v", param.Default)
				}
			default:
				paramStr += fmt.Sprintf(" = %v", param.Default)
			}
		}

		paramStrings[i] = paramStr
	}

	return strings.Join(paramStrings, ", ")
}

// writeTypeValidation generates validation code for parameters
func (t *PythonTranspiler) writeTypeValidation(params []ast.Parameter) error {
	if len(params) == 0 {
		return nil
	}

	t.WriteLine("# Parameter validation")

	for _, param := range params {
		// Skip validation for parameters with default values
		if param.Default != nil {
			continue
		}

		validator, ok := t.GetTypeValidators()[param.Type]
		if !ok {
			t.WriteLine("# No specific validation for type '%s'", param.Type)
			continue
		}

		if err := validator(t, param); err != nil {
			return fmt.Errorf("error validating parameter '%s': %w", param.Name, err)
		}
	}

	return nil
}

// validateStringType validates string parameters
func (t *PythonTranspiler) validateStringType(base BaseTranspiler, param ast.Parameter) error {
	base.WriteLine("if not isinstance(%s, str):", param.Name)
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("raise TypeError(f\"%s must be a string, got {type(%s).__name__}\")",
		param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	return nil
}

// validateNumberType validates numeric parameters
func (t *PythonTranspiler) validateNumberType(base BaseTranspiler, param ast.Parameter) error {
	base.WriteLine("if not isinstance(%s, (int, float)) or isinstance(%s, bool):", param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("raise TypeError(f\"%s must be a number, got {type(%s).__name__}\")",
		param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	return nil
}

// validateIntegerType validates integer parameters
func (t *PythonTranspiler) validateIntegerType(base BaseTranspiler, param ast.Parameter) error {
	base.WriteLine("if not isinstance(%s, int) or isinstance(%s, bool):", param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("raise TypeError(f\"%s must be an integer, got {type(%s).__name__}\")",
		param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	return nil
}

// validateBooleanType validates boolean parameters
func (t *PythonTranspiler) validateBooleanType(base BaseTranspiler, param ast.Parameter) error {
	base.WriteLine("if not isinstance(%s, bool):", param.Name)
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("raise TypeError(f\"%s must be a boolean, got {type(%s).__name__}\")",
		param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	return nil
}

// validateEnumType validates enum parameters
func (t *PythonTranspiler) validateEnumType(base BaseTranspiler, param ast.Parameter) error {
	if len(param.Constraints) == 0 {
		return fmt.Errorf("enum type requires constraints with allowed values")
	}

	values := make([]string, len(param.Constraints))
	for i, c := range param.Constraints {
		values[i] = fmt.Sprintf("%q", c)
	}

	base.WriteLine("%s_valid_values = [%s]", param.Name, strings.Join(values, ", "))

	t.validateStringType(base, param)

	base.WriteLine("if %s not in %s_valid_values:", param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("raise ValueError(f\"%s must be one of {%s_valid_values}\")",
		param.Name, param.Name)
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	return nil
}

// validateFileType validates file parameters
func (t *PythonTranspiler) validateFileType(base BaseTranspiler, param ast.Parameter) error {
	t.validateStringType(base, param)
	base.WriteLine("%s_path = validate_path(%s)", param.Name, param.Name)
	return nil
}

// validateDirectoryType validates directory parameters
func (t *PythonTranspiler) validateDirectoryType(base BaseTranspiler, param ast.Parameter) error {
	return t.validateFileType(base, param)
}

// validateCharacterType validates character parameters
func (t *PythonTranspiler) validateCharacterType(base BaseTranspiler, param ast.Parameter) error {
	return t.validateStringType(base, param)
}

// writeSecurityChecks generates security-related validation code
func (t *PythonTranspiler) writeSecurityChecks(params []ast.Parameter) {
	fileParams := false

	for _, param := range params {
		if param.Type == "file" {
			if !fileParams {
				t.WriteLine("")
				t.WriteLine("# File existence checks")
				fileParams = true
			}

			t.WriteLine("if not is_running_in_docker():")
			t.SetIndentLevel(t.GetIndentLevel() + 1)
			t.WriteLine("if not os.path.isfile(%s_path):", param.Name)
			t.SetIndentLevel(t.GetIndentLevel() + 1)
			t.WriteLine("raise FileNotFoundError(f\"File {%s_path} does not exist\")", param.Name)
			t.SetIndentLevel(t.GetIndentLevel() - 1)
			t.SetIndentLevel(t.GetIndentLevel() - 1)
		} else if param.Type == "directory" {
			if !fileParams {
				t.WriteLine("")
				t.WriteLine("# Directory existence checks")
				fileParams = true
			}

			t.WriteLine("if not is_running_in_docker():")
			t.SetIndentLevel(t.GetIndentLevel() + 1)
			t.WriteLine("if not os.path.isdir(%s_path):", param.Name)
			t.SetIndentLevel(t.GetIndentLevel() + 1)
			t.WriteLine("raise NotADirectoryError(f\"Directory {%s_path} does not exist\")", param.Name)
			t.SetIndentLevel(t.GetIndentLevel() - 1)
			t.SetIndentLevel(t.GetIndentLevel() - 1)
		}
	}
}

// processImplementations handles implementation blocks
func (t *PythonTranspiler) processImplementations(program *ast.Program) error {
	if len(program.Implementations) == 0 {
		t.WriteLine("")
		t.WriteLine("# No implementation blocks found")
		t.WriteLine("raise NotImplementedError(\"No implementation defined for this function\")")
		return nil
	}

	// Process each implementation
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

// handleDockerImplementation generates code for Docker-based implementations
func (t *PythonTranspiler) handleDockerImplementation(base BaseTranspiler, impl *ast.ImplementationBlock, program *ast.Program) error {
	// Extract Docker image
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
			base.WriteLine("%s_abspath = os.path.abspath(%s_path if '%s_path' in locals() else %s)",
				param, param, param, param)
			base.WriteLine("%s_dir = os.path.dirname(%s_abspath)", param, param)
			base.WriteLine("%s_filename = os.path.basename(%s)", param, param)
		}

		// Use first file parameter's directory as main mount point
		base.WriteLine("")
		base.WriteLine("# Main volume mount point")
		base.WriteLine("main_mount_dir = %s_dir", fileParams[0])
	} else {
		// Fallback to current directory
		base.WriteLine("# No file parameters found, using current directory")
		base.WriteLine("main_mount_dir = os.path.abspath(os.getcwd())")
	}

	// Setup execution block with error handling
	base.WriteLine("")
	base.WriteLine("# Execute Docker container with error handling")
	base.WriteLine("try:")
	base.SetIndentLevel(base.GetIndentLevel() + 1)

	// Prepare Docker volumes
	base.WriteLine("# Prepare Docker volumes")
	base.WriteLine("volumes = {}")
	volumes, ok := impl.Fields["volumes"].([]any)
	if ok && len(volumes) > 0 {
		for _, vol := range volumes {
			switch v := vol.(type) {
			case []any:
				if len(v) >= 2 {
					src := fmt.Sprintf("%v", v[0])
					dst := fmt.Sprintf("%v", v[1])

					// Check if src is a parameter reference
					if IsParamReference(src, program.Parameters) {
						base.WriteLine("volumes[%s_dir] = \"%s\"", src, dst)
					} else if src == "parent-folder" || src == "parent_folder" {
						base.WriteLine("volumes[main_mount_dir] = \"%s\"", dst)
					} else {
						base.WriteLine("volumes[\"%s\"] = \"%s\"", src, dst)
					}
				}
			}
		}
	} else {
		// Default volume mapping
		base.WriteLine("volumes[main_mount_dir] = \"/data\"")
	}

	// Prepare environment variables
	base.WriteLine("")
	base.WriteLine("# Prepare environment variables")
	base.WriteLine("env_vars = {}")
	env, ok := impl.Fields["env"].([]any)
	if ok && len(env) > 0 {
		for _, e := range env {
			switch ev := e.(type) {
			case []any:
				if len(ev) >= 2 {
					key := fmt.Sprintf("%v", ev[0])
					val := fmt.Sprintf("%v", ev[1])

					// Check if val is a parameter reference
					if IsParamReference(val, program.Parameters) {
						base.WriteLine("env_vars[\"%s\"] = str(%s)", key, val)
					} else {
						base.WriteLine("env_vars[\"%s\"] = \"%s\"", key, val)
					}
				}
			}
		}
	}

	// Prepare Docker arguments
	base.WriteLine("")
	base.WriteLine("# Prepare Docker arguments")
	base.WriteLine("docker_args = []")
	args, ok := impl.Fields["arguments"].([]any)
	if ok && len(args) > 0 {
		for _, arg := range args {
			argStr := fmt.Sprintf("%v", arg)

			// Skip placeholders
			if argStr == "_" {
				continue
			}

			// Check if it's a parameter reference
			if IsParamReference(argStr, program.Parameters) {
				paramType := GetParamType(argStr, program.Parameters)

				if paramType == "file" || (paramType == "string" && Contains(fileParams, argStr)) {
					// Use filename for file parameters
					base.WriteLine("docker_args.append(%s_filename)", argStr)
				} else if paramType == "boolean" {
					// Convert boolean to flag
					base.WriteLine("if %s:", argStr)
					base.SetIndentLevel(base.GetIndentLevel() + 1)
					base.WriteLine("docker_args.append(\"--true-flag\")")
					base.SetIndentLevel(base.GetIndentLevel() - 1)
				} else {
					base.WriteLine("docker_args.append(str(%s))", argStr)
				}
			} else if strings.HasPrefix(argStr, "\"") || strings.HasPrefix(argStr, "'") {
				// Already a string literal
				base.WriteLine("docker_args.append(%s)", argStr)
			} else {
				// Treat as string
				base.WriteLine("docker_args.append(\"%s\")", argStr)
			}
		}
	}

	// Run the Docker container
	base.WriteLine("")
	base.WriteLine("# Run Docker container")
	base.WriteLine("run_docker(\"%s\", volumes, env_vars, docker_args)", image)

	// Create output directory and return result
	base.WriteLine("")
	base.WriteLine("# Create results directory")
	base.WriteLine("output_dir = os.path.join(main_mount_dir, \"%s_results\")", program.Name)
	base.WriteLine("os.makedirs(output_dir, exist_ok=True)")

	base.WriteLine("")
	base.WriteLine("return Result(status=\"success\", output_dir=output_dir)")

	// Error handling
	base.SetIndentLevel(base.GetIndentLevel() - 1)
	base.WriteLine("except Exception as e:")
	base.SetIndentLevel(base.GetIndentLevel() + 1)
	base.WriteLine("logger.error(f\"Docker execution failed: {str(e)}\")")
	base.WriteLine("return Result(status=\"error\", output_dir=\"\", message=str(e))")
	base.SetIndentLevel(base.GetIndentLevel() - 1)

	return nil
}

// writeEntryPoint adds a main block for direct execution
func (t *PythonTranspiler) writeEntryPoint(program *ast.Program) {
	t.WriteLine("")
	t.WriteLine("")
	t.WriteLine("if __name__ == \"__main__\":")
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	t.WriteLine("import argparse")
	t.WriteLine("")
	t.WriteLine("parser = argparse.ArgumentParser(description=\"%s\")",
		FormatDescription(program.Description))

	var argName string

	// Add arguments for each parameter
	for _, param := range program.Parameters {
		argName = "--" + param.Name
		helpText := param.Description
		if helpText == "" {
			helpText = fmt.Sprintf("Parameter of type '%s'", param.Type)
		}

		switch param.Type {
		case "boolean":
			t.WriteLine("parser.add_argument('%s', action='store_true', help=\"%s\")",
				argName, helpText)
		case "enum":
			if len(param.Constraints) > 0 {
				choices := make([]string, len(param.Constraints))
				for i, c := range param.Constraints {
					choices[i] = fmt.Sprintf("\"%v\"", c)
				}
				choicesStr := strings.Join(choices, ", ")
				t.WriteLine("parser.add_argument('%s', choices=[%s], help=\"%s\")",
					argName, choicesStr, helpText)
			} else {
				t.WriteLine("parser.add_argument('%s', help=\"%s\")", argName, helpText)
			}
		default:
			t.WriteLine("parser.add_argument('%s', help=\"%s\")", argName, helpText)
		}
	}

	t.WriteLine("")
	t.WriteLine("args = parser.parse_args()")
	t.WriteLine("")

	// Call the function with parsed arguments
	t.WriteLine("result = %s(", program.Name)
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	for _, param := range program.Parameters {
		t.WriteLine("%s=args.%s,", param.Name, param.Name)
	}
	t.SetIndentLevel(t.GetIndentLevel() - 1)
	t.WriteLine(")")
	t.WriteLine("")
	t.WriteLine("print(f\"Status: {result.status}\")")
	t.WriteLine("if result.status == \"success\":")
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	t.WriteLine("print(f\"Output directory: {result.output_dir}\")")
	t.SetIndentLevel(t.GetIndentLevel() - 1)
	t.WriteLine("else:")
	t.SetIndentLevel(t.GetIndentLevel() + 1)
	t.WriteLine("print(f\"Error: {result.message}\")")
	t.WriteLine("sys.exit(1)")
	t.SetIndentLevel(t.GetIndentLevel() - 1)
	t.SetIndentLevel(t.GetIndentLevel() - 1)
}
