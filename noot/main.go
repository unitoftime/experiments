// Adapted from: https://github.com/aaronraff/blog-code/blob/master/how-to-write-a-lexer-in-go/lexer.go

package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("input.test")
	if err != nil {
		panic(err)
	}

	tokens := make([]PackedToken, 0)
	lexer := NewLexer(file)
	for {
		pos, tok, lit := lexer.Lex()
		if tok == EOF {
			break
		}

		tokens = append(tokens, PackedToken{pos, tok, lit})
		fmt.Printf("%d:%d\t%s\t%s\n", pos.line, pos.column, tok, lit)
	}

	parser := Parser{}
	tokenList := &Tokens{tokens}
	nodes := parser.Parse(tokenList)
	for _, node := range nodes {
		fmt.Println(node)
	}
}

	// 5 + 4
	// (+ 5 4)
	// NodeExpr(NodeMath(NodeInt(5), NodeInt(4), NodeOperator(PLUS)))
	// NodeExpr(NodeFunc(NodeOperator(PLUS), NodeInt(5), NodeInt(4)))

type PackedToken struct {
	pos Position
	token Token
	str string
}

type Tokens struct {
	list []PackedToken
}

func (t *Tokens) Len() int {
	return len(t.list)
}
func (t *Tokens) Peek() PackedToken {
	return t.list[0]
}
func (t *Tokens) Next() PackedToken {
	token := t.list[0]
	t.list = t.list[1:]
	return token
}


type Node interface {
}

type Parser struct {
}
func (p *Parser) Parse(tokens *Tokens) []Node {
	nodes := make([]Node, 0)
	for tokens.Len() > 0 {
		node := p.ParseDecl(tokens)
		nodes = append(nodes, node)
	}
	return nodes
}

func (p *Parser) ParseDecl(tokens *Tokens) Node {
	next := tokens.Next()
	if next.str == "func" {
		return p.ParseFuncNode(tokens)
	} else if next.str == "return" {
		return p.ParseReturnNode(tokens)
	}

	return nil // TODO fix
}


type FuncNode struct {
	funcName string
	arguments Node
	body Node
}
func (p *Parser) ParseFuncNode(tokens *Tokens) Node {
	next := tokens.Next()
	if next.token != IDENT {
		panic("MUST BE IDENTIFIER")
	}

	// TODO return type

	args := p.ParseArgNode(tokens)
	body := p.Parse(tokens)
	f := FuncNode{
		funcName: next.str,
		arguments: args,
		body: body,
	}
	return f
}

type ReturnNode struct {
	expr Node
}
func (p *Parser) ParseReturnNode(tokens *Tokens) Node {
	r := ReturnNode{
		expr: p.ParseExprNode(tokens),
	}
	return r
}

type Arg struct {
	name string
	kind string
}

type ArgNode struct {
	args []Arg
}
func (p *Parser) ParseArgNode(tokens *Tokens) Node {
	next := tokens.Next()
	if next.token != LPAREN { panic("MUST BE LPAREN") }

	args := ArgNode{make([]Arg, 0)}
	for {
		if tokens.Peek().token == RPAREN { break }

		arg := p.ParseTypedArg(tokens)
		args.args = append(args.args, arg)

		if tokens.Peek().token == COMMA {
			tokens.Next()
		}
	}

	tokens.Next() // Drop the RPAREN

	return args
}

func (p *Parser) ParseTypedArg(tokens *Tokens) Arg {
	name := tokens.Next()
	if name.token != IDENT {
		panic(fmt.Sprintf("MUST BE IDENT: %s", name.str))
	}

	kind := tokens.Next()
	if kind.token != IDENT {
		panic(fmt.Sprintf("MUST BE IDENT: %s", kind.str))
	}

	return Arg{name.str, kind.str}
}

type Operator uint8
const (
	OpNone Operator = iota
	OpAdd
	OpSub
	OpMul
	OpDiv
)

type ExprNode struct {
	// left Node
	// ops []OpExpr
	// right Node
	// op Operator
	ops []Node
}
func (p *Parser) ParseExprNode(tokens *Tokens) Node {
	next := tokens.Next()
	if next.token == LPAREN {
		op := p.ParseExprNode(tokens)
		if tokens.Next().token != RPAREN {
			panic("SHOULD BE RPAREN!!!!")
		}
		return ExprNode{
			ops: []Node{op},
		}
	}

	expr := ExprNode{
		ops: make([]Node, 0),
	}
	for {
		if tokens.Peek().token == RPAREN {
			break
		}

		next := tokens.Next()
		expr.ops = append(expr.ops, UnaryNode{next})
	}

	return expr
}
type UnaryNode struct {
	token PackedToken
}
