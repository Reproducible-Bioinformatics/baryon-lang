package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/reproducible-bioinformatics/baryon-lang/internal/ast"
	"github.com/reproducible-bioinformatics/baryon-lang/internal/lexer"
	"github.com/reproducible-bioinformatics/baryon-lang/internal/parser"
	"github.com/reproducible-bioinformatics/baryon-lang/internal/transpiler"
)

func main() {
	// When check mode is enabled, don't ask for a output file, or a target language.
	check := flag.Bool("check", false, "Check syntax only, do not transpile")
	inputFile := flag.String("input", "", "Input Baryon file (.bala)")
	outputFile := flag.String("output", "", "Output file (default: same name with language-specific extension)")
	langFlag := flag.String("lang", "r",
		fmt.Sprintf("Target language: %s",
			strings.Join(transpiler.GetTranspilerNames(), ", ")))
	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: Input file is required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate target language
	targetLang := strings.ToLower(*langFlag)
	currentTranspiler, err := transpiler.GetTranspiler(targetLang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unsupported language '%s'.", targetLang)
		os.Exit(1)
	}

	// Generate output filename if not provided
	outFile := *outputFile
	if outFile == "" {
		ext := filepath.Ext(*inputFile)
		baseFile := (*inputFile)[0 : len(*inputFile)-len(ext)]
		outFile = baseFile + currentTranspiler.Extension
	}

	fmt.Printf("Reading: %s\n", *inputFile)
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("reading file: %v", err)
	}

	fmt.Println("Parsing Baryon code...")
	program, err := parseProgram(string(data))
	if err != nil {
		log.Fatalf("parsing error: %v", err)
	}

	if *check {
		fmt.Println("✅ Syntax check passed")
		fmt.Print(program.String())
		os.Exit(0)
	}

	// Process and transpile the file
	if err := processFile(outFile, targetLang, currentTranspiler, program); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func processFile(outputPath, lang string,
	currentTranspiler *transpiler.TranspilerDescriptor,
	program *ast.Program,
) error {
	fmt.Printf("Transpiling to %s...\n", currentTranspiler.Display)

	t := currentTranspiler.Initializer()

	code, err := t.Transpile(program)
	if err != nil {
		return fmt.Errorf("transpilation failed: %w", err)
	}

	fmt.Printf("Writing: %s\n", outputPath)
	if err = writeFileSafely(outputPath, []byte(code)); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	fmt.Println("✅ Transpilation completed successfully")
	return nil
}

func parseProgram(source string) (*ast.Program, error) {
	lex := lexer.New(source)
	p := parser.New(lex)
	return p.ParseProgram()
}

// writeFileSafely writes data to a file with appropriate permissions and atomicity
func writeFileSafely(path string, data []byte) error {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	tempFile := path + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}
	return os.Rename(tempFile, path)
}
