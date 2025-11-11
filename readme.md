# Baryon Language

Baryon is a domain-specific language (DSL) and transpiler toolkit for
bioinformatics workflow definition and execution. It allows users to write
workflows in a unified syntax and transpile them into various target languages
and workflow engines, such as Bash, Python, R, Galaxy, Nextflow, and
Streamflow.

## Language Specification

You can find the detailed language specification in the
[SPECIFICATION.md](SPECIFICATION.md) file. It describes the syntax, semantics,
and structure of the Baryon language, including how to define workflows,
metadata, implementation blocks, and parameters.

## Features

- **Unified Workflow DSL:** Write your workflow once, target multiple
platforms.
- **Multiple Transpilers:** Generate code for Bash, Python, R, Galaxy,
Nextflow, and Streamflow.
- **Extensible Architecture:** Easily add new target languages or workflow
engines.
- **Integrated Testing:** Unit and integration tests for core components.

## Getting Started

### Prerequisites

- Go 1.18+ installed on your system (https://golang.org/dl/)

### Installation

Clone the repository:

```sh
git clone https://github.com/yourusername/baryon-lang.git
cd baryon-lang
go build -o baryon-lang main.go
```

### Usage

Transpile a workflow file to a target language:

```sh
./baryon-lang -in examples/enrichment_analysis.bala -target python
```

Supported targets: `bash`, `python`, `r`, `galaxy`, `nextflow`, `streamflow`

## Project Structure

- `internal/ast/` — Abstract syntax tree definitions
- `internal/lexer/` — Lexer for the Baryon DSL
- `internal/parser/` — Parser for the Baryon DSL
- `internal/transpiler/` — Transpilers for supported targets
- `examples/` — Example workflow files
- `main.go` — CLI entry point

## Contributing

Contributions are welcome! Please open issues or submit pull requests.

## License

[MIT](LICENSE)

---

## Tutorial

This documentation will guide you through the concepts, structure, and tooling
of **baryon-lang**, a language for defining reproducible bioinformatics
workflows. The tutorial is based on actual code and examples from the
[baryon-lang
repository](https://github.com/Reproducible-Bioinformatics/baryon-lang), with a
focus on practical insights.

---

## 1. Introduction: What Is baryon-lang?

**baryon-lang** is a domain-specific language (DSL) designed to author, check,
and transpile workflow definitions for bioinformatics, especially in contexts
where reproducibility and platform-independence are critical. Its syntax is
Lisp-inspired (S-expressions), and it is intended to be transpiled to languages
such as R, Python, Nextflow, or bash, maximizing portability and collaboration.

---

## 2. CLI Usage

You can now use the CLI to check or transpile baryon files:

```sh
./baryon-lang -input examples/enrichment_analysis.bala -lang r

./baryon-lang -input another_program.bala -lang galaxy
```

You can also get the latest version of baryon-lang from the GitHub releases
section.

Supported values for `-lang` include: `r`, `python`, `bash`, `nextflow`,
`galaxy` and `streamflow`.

---

## 3. The baryon-lang Syntax: S-Expressions

baryon-lang files are structured as S-expressions (like Lisp), where every
construct is wrapped in parentheses. Here's the opening of a typical file:

```lisp
(bala enrichment_analysis (
  (matrix_file string (desc "Path to the CSV file."))
  ...
))
```

Key S-expression elements:
- **Program node**: `(bala program_name (...))` This will be the name of the
program that will be transpiled. 
- **Parameters**: `(param_name type (desc "..."))` This contains the list of
parameters for you contained function.
- **Implementation blocks**: e.g., `(run_docker ...)` This expressess the core
of your function that will be run inside a Docker container. It will specify
how to interact with the parameters in docker.
- **Descriptions**: `(desc "...")` provides human-readable documentation for
the program or parameters.
- **Outputs**: `(outputs ...)` specifies expected outputs, their types, and
locations.

---

## 4. Defining Parameters

Parameters are defined with a name, type, and optional metadata. Example:

```lisp
(matrix_file string (desc "CSV file of differential expression results."))
(species (enum ("hsapiens" "mmusculus" "dmelanogaster"))
    (desc "Species being analyzed."))
(separator character (desc "Separator character in the table."))
(max_terms number (desc "Max terms in the output."))
```

**Non-obvious detail:**  
- For enums, values are provided as a list of quoted strings.
- The **type** can be `string`, `number`, `character`, `enum`, etc.

---

## 5. Implementation Blocks

The core logic is described in implementation blocks. The most common is
`run_docker`, which specifies a Docker image, volume mappings, and command-line
arguments.

Example:
```lisp
(run_docker
    (image "repbioinfo/singlecelldownstream:latest")
    (volumes (parent_folder "/scratch"))
    (arguments
        "Rscript /home/enrichment_analysis.r"
        matrix_file
        species
        source
        separator
        max_terms
    )
)
```

- `arguments` can mix literals and references to parameters (unquoted).
- Volume source can be a parameter name; the transpiler will resolve it.

---

## 6. Writing a Complete baryon-lang Program

Here is a minimal but complete program, as in
[`examples/enrichment_analysis.bala`](https://github.com/Reproducible-Bioinformatics/baryon-lang/blob/main/examples/enrichment_analysis.bala):

```lisp
(bala enrichment_analysis (
  (matrix_file string (desc "..."))
  (species (enum ("hsapiens" "mmusculus" "dmelanogaster")) (desc "..."))
  (parent_folder string (desc "..."))
  (separator character (desc "..."))
  (max_terms number (desc "..."))
  (run_docker
    (image "repbioinfo/singlecelldownstream:latest")
    (volumes (parent_folder "/scratch"))
    (arguments ...))
  (desc "Process results and perform pathway enrichment.")
  (outputs (scratch directory /scratch))
))
```

- Comments start with `;` and are ignored by the parser.

---

## 7. Syntax Checking

Before transpiling, always check your baryon file for syntax errors:

```sh
./baryon-lang -input myprogram.bala -check
```

The tool will print a summary or detailed error messages (including
line/column).

---

## 8. Transpiling to R, Python, Bash, or Nextflow

To generate code in your target language:

```sh
./baryon-lang -input enrichment_analysis.bala -lang python
```

This produces a `.py` file (or `.R`, `.sh`, `.nf`, etc.) with:
- Function definitions matching the baryon program and its parameters
- Docstrings or comments from `desc`
- Parameter validation (type checks, enum enforcement)
- Secure handling of file paths and Docker calls
- The transpilers generate not only code, but also validation and security
  checks.

---

## 9. Advanced: Enum Constraints and Validation

baryon-lang validates enums at transpile-time and at runtime in the target language:

```lisp
(species (enum ("hsapiens" "mmusculus")) (desc "..."))
```

**Technical detail:**  
- In Python, this becomes a check like `if species not in ['hsapiens', 'mmusculus']: raise ...`
- In R, the function checks `species %in% c("hsapiens", "mmusculus")`

This prevents accidental mis-specification and improves reproducibility.

---

## 10. Output Specification

The `outputs` section describes expected outputs, their types, and locations:

```lisp
(outputs (scratch directory /scratch))
```

**Note:**  
- This is for documentation and for downstream workflow integration. The actual
  code handling output directories is generated based on these specifications.
  Please refer to galaxy output type documentation.

---

## 11. Extending baryon-lang

You can add new parameter types or implementation blocks by editing the Go
source:
- New types: update `internal/ast` and relevant transpilers
- New transpilers: implement the `Transpiler` interface from
  `internal/transpiler/transpiler.go`

---

## 12. Debugging & Testing

- Run `go test ./...` to execute the comprehensive test suite (lexer, parser,
  transpilers).
- Use the parser’s error messages for debugging (they report line/column
  precisely).
## 13. LLM support 
Using the prompt within the file named prompt.txt, users can get help from LLM to generate a scratch version of a bala file based on their script. The bala file needs to be checked and verified by the user. 
