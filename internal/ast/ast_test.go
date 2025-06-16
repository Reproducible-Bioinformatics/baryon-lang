package ast

import (
	"strings"
	"testing"
)

func TestProgramString(t *testing.T) {
	prog := Program{
		NamedBaseNode: NamedBaseNode{
			BaseNode: BaseNode{
				Description: "Test program",
			},
			Name: "myprog",
		},
		Parameters: []Parameter{
			{
				NamedBaseNode: NamedBaseNode{
					BaseNode: BaseNode{
						Description: "A parameter",
					},
					Name: "param1",
				},
				Type:        "string",
				Constraints: []any{"A", "B"},
				Metadata:    map[string]string{"label": "Param 1"},
			},
		},
		Implementations: []ImplementationBlock{
			{
				BaseNode: BaseNode{
					Description: "Run block",
				},
				Name: "run_docker",
				Fields: map[string]any{
					"image": "ubuntu:latest",
					"args":  []string{"echo", "hi"},
				},
			},
		},
		Metadata: map[string]string{"author": "alice"},
	}

	out := prog.String()
	if !strings.Contains(out, "Program: myprog") {
		t.Errorf("missing program name in output")
	}
	if !strings.Contains(out, "Description: Test program") {
		t.Errorf("missing description in output")
	}
	if !strings.Contains(out, "Param: param1") {
		t.Errorf("missing parameter in output")
	}
	if !strings.Contains(out, "Block: run_docker") {
		t.Errorf("missing implementation block in output")
	}
	if !strings.Contains(out, "author: alice") {
		t.Errorf("missing metadata in output")
	}
}

func TestParameterString_Empty(t *testing.T) {
	param := Parameter{
		NamedBaseNode: NamedBaseNode{Name: "p"},
		Type:          "int",
	}
	out := param.String()
	if !strings.Contains(out, "Param: p") {
		t.Errorf("missing param name")
	}
	if !strings.Contains(out, "Type: int") {
		t.Errorf("missing type")
	}
}

func TestImplementationBlockString_EmptyFields(t *testing.T) {
	ib := ImplementationBlock{Name: "block"}
	out := ib.String()
	if !strings.Contains(out, "Block: block") {
		t.Errorf("missing block name")
	}
}

func TestValueString(t *testing.T) {
	v := Value{Literal: 42}
	if got := v.String(); got != "42" {
		t.Errorf("unexpected value string: %s", got)
	}
	v = Value{Identifier: "foo"}
	if got := v.String(); got != "foo" {
		t.Errorf("unexpected identifier string: %s", got)
	}
}
