# Baryon Language

Baryon is a domain-specific language (DSL) and transpiler toolkit for
bioinformatics workflow definition and execution. It allows users to write
workflows in a unified syntax and transpile them into various target languages
and workflow engines, such as Bash, Python, R, Galaxy, Nextflow, and
Streamflow.

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
