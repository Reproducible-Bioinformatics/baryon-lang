package ast

import (
	"fmt"
)

// BaseNode represents the common fields for all AST nodes.
type BaseNode struct {
	fmt.Stringer
	Type        string
	Description string
}

// NamedBaseNode represents a BaseNode with a name field.
type NamedBaseNode struct {
	BaseNode
	Name string
}

// Program represents the root of the Abstract Syntax Tree, defining a program.
type Program struct {
	NamedBaseNode
	Parameters     []Parameter
	Implementation Implementation
}

// ParameterReference represents a reference to a parameter within the program.
type ParameterReference struct {
	NamedBaseNode
}

// Parameter defines a parameter for the program.
type Parameter struct {
	NamedBaseNode
	Constraints []any
	Default     any
}

// Implementation specifies how the program is implemented (using Docker).
type Implementation struct {
	BaseNode
	Image     string
	Volumes   []Volume
	Arguments []any
}

// Volume defines a Docker volume mount.
type Volume struct {
	BaseNode
	HostPath      string
	ContainerPath string
}
