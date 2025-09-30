package ast

import (
	"bytes"
	"fmt"
)

// BaseNode represents the common fields for all AST nodes.
type BaseNode struct {
	fmt.Stringer
	Description string
}

// NamedBaseNode represents a BaseNode with a name field.
type NamedBaseNode struct {
	BaseNode
	Name string
}

// Program represents the root of the Abstract Syntax Tree.
type Program struct {
	NamedBaseNode
	Parameters      []Parameter
	Implementations []ImplementationBlock
	Metadata        map[string]string
	Outputs         []OutputBlock
}

func (p Program) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Program: %s\n", p.Name))
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("\tDescription: %s\n", p.Description))
	}
	if len(p.Metadata) > 0 {
		buf.WriteString("\tMetadata:\n")
		for k, v := range p.Metadata {
			buf.WriteString(fmt.Sprintf("\t\t%s: %s\n", k, v))
		}
	}
	if len(p.Parameters) > 0 {
		buf.WriteString("\tParameters:\n")
		for _, parameter := range p.Parameters {
			buf.WriteString(parameter.String())
		}
	}
	if len(p.Implementations) > 0 {
		buf.WriteString("\tImplementations:\n")
		for _, impl := range p.Implementations {
			buf.WriteString(impl.String())
		}
	}
	if len(p.Outputs) > 0 {
		buf.WriteString("\tOutputs:\n")
		for _, output := range p.Outputs {
			buf.WriteString(output.String())
		}
	}
	return buf.String()
}

// Parameter defines a parameter for the program.
type Parameter struct {
	NamedBaseNode
	Type        string
	Constraints []any // For enum type
	Default     any
	Metadata    map[string]string // extensible (e.g., label)
}

func (p Parameter) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("\t\tParam: %s\n", p.Name))
	buf.WriteString(fmt.Sprintf("\t\t\tType: %s\n", p.Type))
	if len(p.Constraints) > 0 {
		buf.WriteString(fmt.Sprintf("\t\t\tConstraints: %v\n", p.Constraints))
	}
	if p.Description != "" {
		buf.WriteString(fmt.Sprintf("\t\t\tDescription: %s\n", p.Description))
	}
	if len(p.Metadata) > 0 {
		buf.WriteString("\t\t\tMetadata:\n")
		for k, v := range p.Metadata {
			buf.WriteString(fmt.Sprintf("\t\t\t\t%s: %s\n", k, v))
		}
	}
	return buf.String()
}

// ImplementationBlock is a generic node for any implementation section
type ImplementationBlock struct {
	BaseNode
	Name   string         // e.g., "run_docker"
	Fields map[string]any // Holds fields like "image", "volumes", "arguments" and their values
}

func (ib ImplementationBlock) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("\t\tBlock: %s\n", ib.Name))
	if len(ib.Fields) > 0 {
		buf.WriteString("\t\t\tFields:\n")
		for k, v := range ib.Fields {
			buf.WriteString(fmt.Sprintf("\t\t\t\t%s: %v\n", k, v))
		}
	}
	return buf.String()
}

// Represents a value which could be a literal or an identifier reference
type Value struct {
	Literal    any    // string, number, bool, special like "_"
	Identifier string // reference to a parameter, etc.
}

func (v Value) String() string {
	if v.Identifier != "" {
		return v.Identifier
	}
	return fmt.Sprintf("%#v", v.Literal)
}

// OutputBlock defines an output specification for the program.
type OutputBlock struct {
	NamedBaseNode
	Format string // e.g., "json", "tsv"
	Path   string // path to the output file
}

func (ob OutputBlock) String() string {
	var buf bytes.Buffer
	buf.WriteString("\t\tOutput:\n")
	if ob.Format != "" {
		buf.WriteString(fmt.Sprintf("\t\t\tFormat: %s\n", ob.Format))
	}
	buf.WriteString(fmt.Sprintf("\t\t\tPath: %s\n", ob.Path))
	if ob.Description != "" {
		buf.WriteString(fmt.Sprintf("\t\t\tDescription: %s\n", ob.Description))
	}
	return buf.String()
}
