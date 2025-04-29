package parser

import (
	"errors"
	"fmt"
	"iter"
	"strings"

	"github.com/reproducible-bioinformatics/baryon-lang/internal/ast"
	"github.com/reproducible-bioinformatics/baryon-lang/internal/lexer"
)

type Parser struct {
	lexer        *lexer.Lexer
	nextToken    func() (lexer.Token, bool)
	stopIter     func()
	currentToken lexer.Token
	peekToken    lexer.Token
	errors       []string
}

// Structure to represent an S-expression node (for intermediate parsing)
type SExpr struct {
	Token    lexer.Token
	Children []*SExpr
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}
	p.nextToken, p.stopIter = iter.Pull(l.Token())
	p.advance() // Set currentToken
	p.advance() // Set peekToken
	return p
}

func (p *Parser) advance() {
	p.currentToken = p.peekToken
	var ok bool
	for {
		p.peekToken, ok = p.nextToken()
		if !ok || p.peekToken.Type != lexer.TOKEN_COMMENT {
			break
		}
	}
	if !ok {
		p.peekToken = lexer.Token{Type: lexer.TOKEN_EOF}
	}
}

func (p *Parser) ParseProgram() (*ast.Program, error) {
	defer p.stopIter()

	// Parse the entire file into an S-expression tree
	root, err := p.parseSExpr()
	if err != nil {
		return nil, err
	}

	// Transform the S-expression tree into an AST
	program, err := p.sExprToAST(root)
	if err != nil {
		return nil, err
	}

	if len(p.errors) > 0 {
		return nil, p.getError()
	}

	return program, nil
}

// Parse the current input into an S-expression tree
func (p *Parser) parseSExpr() (*SExpr, error) {
	// Skip to the first opening parenthesis
	for p.currentToken.Type != lexer.TOKEN_LPAREN && p.currentToken.Type != lexer.TOKEN_EOF {
		p.advance()
	}

	if p.currentToken.Type == lexer.TOKEN_EOF {
		p.addError("unexpected end of input before program definition")
		return nil, p.getError()
	}

	// Parse the program S-expression
	return p.parseSExprNode()
}

// Parse a single S-expression node and its children
func (p *Parser) parseSExprNode() (*SExpr, error) {
	node := &SExpr{
		Token:    p.currentToken,
		Children: []*SExpr{},
	}

	// If this is an opening parenthesis, parse its contents
	if p.currentToken.Type == lexer.TOKEN_LPAREN {
		p.advance() // Consume the opening parenthesis

		// Parse all child nodes until we hit the closing parenthesis
		for p.currentToken.Type != lexer.TOKEN_RPAREN && p.currentToken.Type != lexer.TOKEN_EOF {
			if p.currentToken.Type == lexer.TOKEN_LPAREN {
				// Parse nested S-expression
				child, err := p.parseSExprNode()
				if err != nil {
					return nil, err
				}
				node.Children = append(node.Children, child)
			} else {
				// Add token as a leaf node
				leaf := &SExpr{
					Token:    p.currentToken,
					Children: []*SExpr{},
				}
				node.Children = append(node.Children, leaf)
				p.advance() // Consume the token
			}
		}

		if p.currentToken.Type == lexer.TOKEN_RPAREN {
			p.advance() // Consume the closing parenthesis
		} else {
			p.addError("missing closing parenthesis in S-expression")
			return nil, p.getError()
		}
	} else {
		// For non-parenthesis tokens, just consume and return
		p.advance()
	}

	return node, nil
}

// Transform an S-expression tree into an AST
func (p *Parser) sExprToAST(root *SExpr) (*ast.Program, error) {
	if len(root.Children) < 3 {
		p.addError("invalid program structure: not enough elements")
		return nil, p.getError()
	}

	// First child should be 'bala'
	if root.Children[0].Token.Type != lexer.TOKEN_IDENTIFIER ||
		root.Children[0].Token.Literal != "bala" {
		p.addError("program must start with 'bala'")
		return nil, p.getError()
	}

	// Second child should be the program name
	if root.Children[1].Token.Type != lexer.TOKEN_IDENTIFIER {
		p.addError("invalid program name")
		return nil, p.getError()
	}

	program := &ast.Program{
		NamedBaseNode: ast.NamedBaseNode{
			Name: root.Children[1].Token.Literal,
		},
		Parameters:      []ast.Parameter{},
		Implementations: []ast.ImplementationBlock{},
		Metadata:        make(map[string]string),
	}

	// Third child should be the program body
	if len(root.Children) < 3 {
		p.addError("program body is empty")
		return nil, p.getError()
	}

	programBody := root.Children[2]

	// Process each element in the program body
	for _, child := range programBody.Children {
		if len(child.Children) == 0 {
			continue // Skip empty nodes
		}

		// Check the first element to determine what kind of node this is
		firstElement := child.Children[0]

		if firstElement.Token.Type != lexer.TOKEN_IDENTIFIER {
			p.addError(fmt.Sprintf("unexpected token %s in program body", firstElement.Token.Type))
			continue
		}

		switch firstElement.Token.Literal {
		case "desc":
			// Program description
			if len(child.Children) > 1 && child.Children[1].Token.Type == lexer.TOKEN_STRING {
				program.Description = child.Children[1].Token.Literal
			}
		case "run_docker":
			// Implementation block
			impl := p.parseImplementationBlockSExpr(child)
			program.Implementations = append(program.Implementations, impl)
		default:
			// Must be a parameter definition
			param := p.parseParameterSExpr(child)
			program.Parameters = append(program.Parameters, param)
		}
	}

	return program, nil
}

// Parse a parameter definition from an S-expression
func (p *Parser) parseParameterSExpr(node *SExpr) ast.Parameter {
	if len(node.Children) == 0 {
		return ast.Parameter{} // Return empty parameter if node is invalid
	}

	// First child is parameter name
	paramName := node.Children[0].Token.Literal

	param := ast.Parameter{
		NamedBaseNode: ast.NamedBaseNode{
			Name: paramName,
		},
		Metadata: make(map[string]string),
	}

	// Second child should be type or enum
	if len(node.Children) > 1 {
		// Handle different parameter type formats
		if node.Children[1].Token.Type == lexer.TOKEN_IDENTIFIER {
			if node.Children[1].Token.Literal == "enum" {
				// Handle "enum" followed by values list
				param.Type = "enum"

				// Process enum values starting from the third child
				if len(node.Children) > 2 {
					for i := 2; i < len(node.Children); i++ {
						child := node.Children[i]
						// Skip desc nodes
						if len(child.Children) > 0 && child.Children[0].Token.Literal == "desc" {
							continue
						}

						// Process string values or nested values
						if child.Token.Type == lexer.TOKEN_STRING {
							// Direct string value
							param.Constraints = append(param.Constraints, child.Token.Literal)
						} else if len(child.Children) > 0 {
							// Values in a nested list
							for _, valueNode := range child.Children {
								if valueNode.Token.Type == lexer.TOKEN_STRING {
									param.Constraints = append(param.Constraints, valueNode.Token.Literal)
								}
							}
						}
					}
				}
			} else {
				// Simple type like "string", "number", etc.
				param.Type = node.Children[1].Token.Literal
			}
		} else if len(node.Children[1].Children) > 0 {
			// Handle "(enum (...)" format
			if node.Children[1].Children[0].Token.Literal == "enum" {
				param.Type = "enum"

				// Process nested enum values
				for i := 1; i < len(node.Children[1].Children); i++ {
					enumValueNode := node.Children[1].Children[i]
					if len(enumValueNode.Children) > 0 {
						for _, value := range enumValueNode.Children {
							if value.Token.Type == lexer.TOKEN_STRING {
								param.Constraints = append(param.Constraints, value.Token.Literal)
							}
						}
					}
				}
			}
		}
	}

	// Process metadata blocks
	for i := 2; i < len(node.Children); i++ {
		metaNode := node.Children[i]

		// Check if it's a metadata block (should start with identifier)
		if len(metaNode.Children) > 0 && metaNode.Children[0].Token.Type == lexer.TOKEN_IDENTIFIER {
			keyword := metaNode.Children[0].Token.Literal

			if keyword == "desc" && len(metaNode.Children) > 1 {
				if metaNode.Children[1].Token.Type == lexer.TOKEN_STRING {
					desc := metaNode.Children[1].Token.Literal
					param.Description = desc
					param.Metadata["desc"] = desc
				}
			} else if len(metaNode.Children) > 1 {
				// Other metadata
				param.Metadata[keyword] = metaNode.Children[1].Token.Literal
			}
		}
	}

	return param
}

// Parse an implementation block from an S-expression
func (p *Parser) parseImplementationBlockSExpr(node *SExpr) ast.ImplementationBlock {
	block := ast.ImplementationBlock{
		Name:   node.Children[0].Token.Literal,
		Fields: make(map[string]any),
	}

	// Process each field in the implementation block
	for i := 1; i < len(node.Children); i++ {
		fieldNode := node.Children[i]

		if len(fieldNode.Children) == 0 {
			continue // Skip empty fields
		}

		// If field has a name, process as a named field
		if fieldNode.Children[0].Token.Type == lexer.TOKEN_IDENTIFIER {
			fieldName := fieldNode.Children[0].Token.Literal

			if fieldName == "image" || fieldName == "command" {
				// Simple string value fields
				if len(fieldNode.Children) > 1 && fieldNode.Children[1].Token.Type == lexer.TOKEN_STRING {
					block.Fields[fieldName] = fieldNode.Children[1].Token.Literal
				}
			} else if fieldName == "volumes" {
				// Volumes with nested key-value pairs
				volumes := []any{}

				// Process each volume definition
				for j := 1; j < len(fieldNode.Children); j++ {
					volumeNode := fieldNode.Children[j]

					if len(volumeNode.Children) >= 2 {
						// Create a key-value pair from first two children
						key := volumeNode.Children[0].Token.Literal
						value := volumeNode.Children[1].Token.Literal

						// Store as an array to preserve order
						volumes = append(volumes, []any{key, value})
					}
				}

				block.Fields[fieldName] = volumes
			} else if fieldName == "arguments" {
				// Arguments list
				args := []any{}

				// Direct argument values
				for j := 1; j < len(fieldNode.Children); j++ {
					argNode := fieldNode.Children[j]

					// Can be string or identifier
					args = append(args, argNode.Token.Literal)
				}

				block.Fields[fieldName] = args
			} else {
				// Generic field handling
				if len(fieldNode.Children) > 1 {
					block.Fields[fieldName] = fieldNode.Children[1].Token.Literal
				} else {
					block.Fields[fieldName] = nil
				}
			}
		}
	}

	return block
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("Line %d, Column %d: %s",
		p.currentToken.Line, p.currentToken.Column, msg))
}

func (p *Parser) getError() error {
	return errors.New(strings.Join(p.errors, "\n"))
}
