# Baryon Language Specification

The Baryon Language is a domain-specific language (DSL) for defining
bioinformatics workflows. It enables users to describe workflows in a unified
syntax, which can be transpiled into multiple target languages and workflow
engines.

The keywords "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this
document are to be interpreted as described in BCP 14
[RFC2119](https://www.ietf.org/rfc/rfc2119.txt)
[RFC8174](https://www.ietf.org/rfc/rfc8174.txt) when, and only when, they
appear in all capitals, as shown here.

## Syntax

### General

- The language syntax is based on S-expressions (parenthesized lists).
- All forms MUST be enclosed in balanced parentheses.
- Identifiers, keywords, and literals (strings, numbers, booleans) are supported.

### Program Structure


A Baryon program MUST be defined as a single S-expression with the following
structure:

```
(bala <program_name>
  (
    <metadata>*
    <implementation_block>*
    <parameter>*
  )
)
```

- The top-level form MUST start with the identifier `bala`.
- `<program_name>` MUST be a valid identifier.
- The body MUST be a list containing zero or more metadata, implementation
blocks, and parameter definitions.

### Metadata

- Metadata blocks MAY be included in the program body.
- The `(desc <string>)` form SHOULD be used to provide a program description.
- Additional metadata MAY be specified as `(key value)` pairs.

### Implementation Blocks

- Implementation blocks define workflow execution details.
- At least one implementation block SHOULD be present.
- The implementation block type (e.g., `run_docker`) MUST be the first element
of the block.
- Supported fields for `run_docker` implementation blocks include:
  - `(image <string>)` (REQUIRED): The Docker image to use.
  - `(command <string>)` (OPTIONAL): The command to execute.
  - `(volumes ((<host_path> <container_path>) ...))` (OPTIONAL): Volume
  mappings.
  - `(env ((<key> <value>) ...))` (OPTIONAL): Environment variables.
  - `(arguments (<arg1> <arg2> ...))` (OPTIONAL): Command-line arguments.

### Parameters

- Parameters MUST be defined as S-expressions in the form:
```
(<name> <type> (<desc <string>>) ...)
```
- `<name>` MUST be a valid identifier.
- `<type>` MUST be one of: `string`, `number`, `integer`, `boolean`, `file`,
`directory`, `character`, `enum`.
- The `(desc <string>)` metadata SHOULD be provided for each parameter.
- The `(default <value>)` metadata MAY be provided to specify a default value.
- Enum parameters MUST specify allowed values using the `(enum (<value1>
<value2> ...))` form.

#### Example

```
(param1 string (desc "A string param"))
(param2 (enum ("A" "B" "C")) (desc "enum param"))
```

### Enum Parameters

- If a parameter type is `enum`, it MUST specify a non-empty list of allowed
values.
- Enum values MUST be strings.

## Constraints and Error Handling

- All parentheses MUST be balanced.
- The top-level form MUST start with `bala`; otherwise, the program is invalid.
- Enum parameters without allowed values MUST cause a parse error.
- If a required field (such as `image` in `run_docker`) is missing,
transpilation MUST fail with an error.
- Unknown or unsupported types SHOULD result in a warning or error.


## Extensibility

- New implementation block types and parameter types MAY be added in future
versions.
- Implementations and parameters not recognized by the transpiler SHOULD be
ignored or cause a warning, depending on the context.

## Comments

The language supports comments. Comments MUST start with a semicolon
(`;`) and continue to the end of the line. Comments can be placed on their own
line or at the end of a line after code.

## Example Program

```
(bala enrichment_analysis
  (
    (desc "Gene set enrichment analysis workflow")
    (run_docker
      (image "biocontainers/enrichment:latest")
      (command "run_enrichment")
      (volumes (("input" "/data/input") ("output" "/data/output")))
      (env (("MODE" "fast")))
      (arguments ("--input" param1 "--output" param2))
    )
    (param1 file (desc "Input file"))
    (param2 string (desc "Output prefix"))
    (param3 (enum ("A" "B" "C")) (desc "Analysis mode"))
  )
)
```
