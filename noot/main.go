// Adapted from: https://github.com/aaronraff/blog-code/blob/master/how-to-write-a-lexer-in-go/lexer.go

// Run this: go run . && dot -Tpdf output.dot > output.pdf

package main

import (
	"fmt"
	"os"
	"io/fs"
	"bytes"
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

		tokens = append(tokens, PackedToken{pos, tok, lit})
		fmt.Printf("%d:%d\t%s\t%s\n", pos.line, pos.column, tok, lit)
		if tok == EOF {
			break
		}
	}

	parser := Parser{}
	tokenList := &Tokens{tokens}
	nodes := parser.ParseFile("input_test", tokenList) // TODO - token to represent file start?
	// for _, node := range nodes {
	// 	fmt.Println(node)
	// }
	buf := bytes.Buffer{}
	nodes.WalkGraphviz("", &buf)

	err = os.WriteFile("output.dot", buf.Bytes(), fs.ModePerm)
	if err != nil {
		panic(err)
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
	WalkGraphviz(string, *bytes.Buffer)
}

type FileNode struct {
	filename string
	nodes []Node
}
func (n *FileNode) WalkGraphviz(prev string, buf *bytes.Buffer) {
	fmt.Println("FileNode")
	buf.WriteString("strict digraph {\n")
	buf.WriteString("node [shape=box]\n")
	buf.WriteString(n.filename + "\n")
	for i := range n.nodes {
		n.nodes[i].WalkGraphviz(n.filename, buf)
	}
	buf.WriteString("\n}")
}
type FuncNode struct {
	funcName string
	arguments Node
	body Node
}
func (n *FuncNode) WalkGraphviz(prev string, buf *bytes.Buffer) {
	fmt.Println(n.funcName)

	buf.WriteString(fmt.Sprintf("%s -> %s\n", prev, n.funcName))

	n.arguments.WalkGraphviz(n.funcName, buf)
	n.body.WalkGraphviz(n.funcName, buf)
}
type CurlyScope struct {
	nodes []Node
}
func (n *CurlyScope) WalkGraphviz(prev string, buf *bytes.Buffer) {
	fmt.Println("CurlyScope")
	for i := range n.nodes {
		n.nodes[i].WalkGraphviz(prev, buf)
	}
}

type ReturnNode struct {
	expr Node
}
func (n *ReturnNode) WalkGraphviz(prev string, buf *bytes.Buffer) {
	nodeName := prev+"_Return" // todo - line number to disambiguate?
	buf.WriteString(fmt.Sprintf("%s -> %s\n", prev, nodeName))
	n.expr.WalkGraphviz(nodeName, buf)
}

type Arg struct {
	name string
	kind string
}
type ArgNode struct {
	args []Arg
}
func (n *ArgNode) WalkGraphviz(prev string, buf *bytes.Buffer) {
	nodeName := prev+"Args"
	label := "[label=\"Args: "
	for i := range n.args {
		fmt.Println(n.args[i].name, n.args[i].kind)
		label = label + n.args[i].name + " " + n.args[i].kind + ", "
	}
	label = label + "\"];"

	buf.WriteString(fmt.Sprintf("%s %s\n", nodeName, label))
	buf.WriteString(fmt.Sprintf("%s -> %s\n", prev, nodeName))
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
	ops []Node
}
func (n *ExprNode) WalkGraphviz(prev string, buf *bytes.Buffer) {
	expr := prev+"Expr"
	buf.WriteString(fmt.Sprintf("%s -> %s\n", prev, expr))
	for i := range n.ops {
		n.ops[i].WalkGraphviz(expr, buf)
	}
}

type UnaryNode struct {
	index int
	token PackedToken
}
func (n *UnaryNode) WalkGraphviz(prev string, buf *bytes.Buffer) {
	expr := fmt.Sprintf("_%d", n.index) + prev + n.token.token.String()

	if n.token.token == IDENT || n.token.token == INT {
		label := fmt.Sprintf("[label=\"%s\"];", n.token.str)
		buf.WriteString(fmt.Sprintf("%s %s\n", expr, label))
	} else {
		label := fmt.Sprintf("[label=\"%s\"];", n.token.token.String())
		buf.WriteString(fmt.Sprintf("%s %s\n", expr, label))
	}
	buf.WriteString(fmt.Sprintf("%s -> %s\n", prev, expr))
	fmt.Println(n.token)
}

// --------------------------------------------------------------------------------
// - Parser
// --------------------------------------------------------------------------------
func (p *Parser) ParseFile(name string, tokens *Tokens) *FileNode {
	return &FileNode{
		name,
		p.ParseTil(tokens, EOF),
	}
}

type Parser struct {
}
func (p *Parser) ParseTil(tokens *Tokens, stopToken Token) []Node {
	nodes := make([]Node, 0)
	for tokens.Len() > 0 {

		node := p.ParseDecl(tokens)
		if node != nil {
			nodes = append(nodes, node)
		} else {
			next := tokens.Next()
			if next.token == stopToken {
				return nodes
			} else {
				if next.token != SEMI {
					panic(fmt.Sprintf("Expected %s - Got: %s", stopToken.String(), next.str))
				}
			}
		}
	}
	return nodes
}

func (p *Parser) ParseDecl(tokens *Tokens) Node {
	next := tokens.Peek()

	if next.str == "func" {
		tokens.Next()
		return p.ParseFuncNode(tokens)
	} else if next.str == "return" {
		tokens.Next()
		return p.ParseReturnNode(tokens)
	}

	return nil // TODO fix
}

// Parsing functions


func (p *Parser) ParseFuncNode(tokens *Tokens) Node {
	next := tokens.Next()
	if next.token != IDENT {
		panic("MUST BE IDENTIFIER")
	}

	// TODO return type

	args := p.ParseArgNode(tokens)
	body := p.ParseCurlyScope(tokens)
	f := FuncNode{
		funcName: next.str,
		arguments: args,
		body: body,
	}

	return &f
}

func (p *Parser) ParseCurlyScope(tokens *Tokens) Node {
	next := tokens.Next()
	if next.token != LBRACE {
		panic("MUST BE LBRACE")
	}

	body := p.ParseTil(tokens, RBRACE)

	return &CurlyScope{body}
}


func (p *Parser) ParseReturnNode(tokens *Tokens) Node {
	r := ReturnNode{
		expr: p.ParseExprNode(tokens),
	}
	return &r
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

	return &args
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

func (p *Parser) ParseExprNode(tokens *Tokens) Node {
	peek := tokens.Peek()
	if peek.token == LPAREN {
		tokens.Next()
		// Case where we have a subexpression
		op := p.ParseExprNode(tokens)
		// if tokens.Next().token != RPAREN {
		// 	panic("SHOULD BE RPAREN!!!!")
		// }
		return &ExprNode{
			ops: []Node{op},
		}
	}

	expr := ExprNode{
		ops: make([]Node, 0),
	}
	// Case where we have a (potentially long) flat expression
	idx := -1
	for {
		idx++
		if tokens.Peek().token == RPAREN {
			tokens.Next()
			break
		}
		if tokens.Peek().token == SEMI {
			tokens.Next()
			break
		}
		if tokens.Peek().token == LPAREN {
			expr.ops = append(expr.ops, p.ParseExprNode(tokens))
			continue
		}

		next := tokens.Next()
		expr.ops = append(expr.ops, &UnaryNode{idx, next})
	}

	return &expr
}
