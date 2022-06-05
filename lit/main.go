// This is an experiment in creating literate go files
package main

// I was curious if I could create something that parses this repo, and generates blog-style pages representing all of the test cases and code, documented via comment blogs. Let's give it a shot. I think I'll have to do some Abstract-Syntax-Tree (AST) crawling, and then template the data into either markdown, or HTML templates. Not sure which yet. I may have to come back here and see if I there is a markdown -> HTML converter package in Go that I'd be willing to use.

import (
	// First we need to import some packages that can do file reading and writing
	"bytes"
	// "fmt"
	"io/fs"
	"os"

	// Import some packages that do AST crawling and formatting the output
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	// I also probably at least one package to convert to HTML. I ended up going with markdown -> HTML
	"github.com/russross/blackfriday/v2" // Test comment

	// We will also embed some other template-type files into this binary
	_ "embed"
)

// So we are going to define an HTML header file which is used to generate the top of an HTML page

//go:embed header.html
var htmlHeader string

// We will also add an HTML footer file to cap off the other end.

//go:embed footer.html
var htmlFooter string

// Let's have a main function that just wraps some high level logic that we will write below
func main() {
	generatePackage(".")
}

/*
This is a block comment do describe this function. This function does
the bulk of the work, deciding which nodes to print.
*/
func generatePackage(dir string) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for _, pkg := range packages {
		tokenStart := pkg.Pos()
		bv := BlogVisitor{
			buf: &bytes.Buffer{},
			pkg: pkg,
			fset: fset,
			lastCommentPos: &tokenStart,
		}
		ast.Walk(bv, pkg)

		bv.Output("out2.html")
	}

/*
	for k, pkg := range packages {
		fmt.Println(k, pkg.Name)

		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf("# %s", pkg.Name))
		for _, file := range pkg.Files {
			lastCommentPos := file.Pos()
			cmap := ast.NewCommentMap(fset, file, file.Comments)
			// file.Comments = cmap.Filter(file).Comments()

			// ast.Print(fset, file)

			for _, decl := range file.Decls {
				// Add any comments that are in this range from the comment map into the output file
				// Notably, file.Comments is already in order for us so we don't need to sort it!
				for _, cgroup := range file.Comments {
					for _, c := range cgroup.List {

						// Our current comment needs to be after the `lastCommentPos` and before the next declaration pos
						if c.Slash > lastCommentPos && c.Slash < decl.Pos() {
							buf.WriteString(fmt.Sprintf("\n\n%s\n\n", c.Text))
						}
					}
				}
				// We set the lastCommentPos to the end of the current declaration.
				// This is because we handle comments in the declaration a different way,
				// so we want to skip those too.
				lastCommentPos = decl.End()

				switch d := decl.(type) {
				case *ast.FuncDecl:
					cgroups := cmap.Filter(d).Comments()
					buf.WriteString("\n" + d.Doc.Text() + "\n") // line comment
					formatFunc(&buf, fset, *d, cgroups)

				case *ast.GenDecl:
					cgroups := cmap.Filter(d).Comments()
					buf.WriteString("\n" + d.Doc.Text() + "\n")
					formatGen(&buf, fset, *d, cgroups)
				}
			}
		}

		markdown := blackfriday.Run(buf.Bytes())

		err := os.WriteFile("output.html", []byte(htmlHeader), fs.ModePerm)
		if err != nil {
			panic(err)
		}

		file, err := os.OpenFile("output.html", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			panic(err)
		}

		_, err = file.Write(markdown)
		if err != nil {
			panic(err)
		}

		_, err = file.Write([]byte(htmlFooter))
		if err != nil {
			panic(err)
		}
	}
*/
}

func formatFunc(buf *bytes.Buffer, fset *token.FileSet, decl ast.FuncDecl, cGroups []*ast.CommentGroup) {
	decl.Doc = nil // nil the Doc field so that we don't print it
	buf.WriteString("\n```go\n")
	commentedNode := printer.CommentedNode{
		Node:     &decl,
		Comments: cGroups,
	}
	formatNode(buf, fset, &commentedNode)
	buf.WriteString("\n```\n")
}

func formatGen(buf *bytes.Buffer, fset *token.FileSet, decl ast.GenDecl, cGroups []*ast.CommentGroup) {
	// decl.Doc = nil // nil the Doc field so that we don't print it
	commentedNode := printer.CommentedNode{
		Node:     &decl,
		Comments: cGroups,
	}

	if decl.Tok == token.IMPORT {
		buf.WriteString("\n```go\n")
		formatNode(buf, fset, &commentedNode)
		buf.WriteString("\n```\n")
	} else if decl.Tok == token.TYPE {
		buf.WriteString("\n```go\n")
		formatNode(buf, fset, &commentedNode)
		buf.WriteString("\n```\n")
	} else if decl.Tok == token.CONST || decl.Tok == token.VAR {
		buf.WriteString("\n```go\n")
		formatRawNode(buf, fset, &commentedNode)
		buf.WriteString("\n```\n")
	}
}

func formatRawNode(buf *bytes.Buffer, fset *token.FileSet, node any) {
	config := printer.Config{
		Mode: printer.RawFormat,
		Tabwidth: 2,
	}
	err := config.Fprint(buf, fset, node)
	if err != nil {
		panic(err)
	}
}

// node can be either a commented node or a node supported by fprintf
func formatNode(buf *bytes.Buffer, fset *token.FileSet, node any) {
	config := printer.Config{
		Mode:     printer.UseSpaces,
		Tabwidth: 2,
	}
	err := config.Fprint(buf, fset, node)
	if err != nil {
		panic(err)
	}
}

type BlogVisitor struct {
	buf *bytes.Buffer
	pkg *ast.Package
	fset *token.FileSet
	file *ast.File
	cmap ast.CommentMap

	lastCommentPos *token.Pos
}
func (v BlogVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil { return nil }

	// Keep searching for something printable if we hit any of these cases
	_, ok := node.(*ast.Package)
	if ok { return v } // TODO - print doc headers?

	file, ok := node.(*ast.File)
	if ok {
		// v.buf.WriteString("\n" + file.Doc.Text() + "\n")
		// if file.Doc != nil {
		// 	// v.handleComments(file.Doc.End(), file.Doc)
		// 	*v.lastCommentPos = file.Doc.End()
		// }

		v.file = file
		v.cmap = ast.NewCommentMap(v.fset, file, file.Comments)
		return v
	}

	f, ok := node.(*ast.FuncDecl)
	if ok {
		// Handle comments
		// if f.Doc != nil {
		// 	v.handleComments(f.Doc.Pos())
		// } else {
		// 	v.handleComments(f.Pos())
		// }
		v.handleComments(f.Pos())
		*v.lastCommentPos = node.End()

		// Handle function case
		cgroups := v.cmap.Filter(f).Comments()
		// v.buf.WriteString("\n" + f.Doc.Text() + "\n")
		formatFunc(v.buf, v.fset, *f, cgroups)
		// fmt.Println(f.Name)

		return nil
	}

	gen, ok := node.(*ast.GenDecl)
	if ok {
		// Handle comments
		// if gen.Doc != nil {
		// 	v.handleComments(gen.Doc.Pos())
		// } else {
		// 	v.handleComments(gen.Pos())
		// }
		v.handleComments(gen.Pos())
		*v.lastCommentPos = node.End()

		// Handle node
		cgroups := v.cmap.Filter(gen).Comments()
		// v.buf.WriteString("\n" + gen.Doc.Text() + "\n")
		formatGen(v.buf, v.fset, *gen, cgroups)

		return nil
	}

	// If all else fails, then keep looking
	return v
}

// from represents the position where we want to start checking comments
// to represents the position where we want to stop checking comments
func (v *BlogVisitor) handleComments(nextPos token.Pos) {
	// Try to printout any comments that are next in line
	if v.file != nil {
		for _, cgroup := range v.file.Comments {
			if cgroup.Pos() > *v.lastCommentPos && cgroup.Pos() < nextPos {
				v.buf.WriteString(cgroup.Text())
			}

			// for _, c := range cgroup.List {
			// 	fmt.Println(c.Slash, v.lastCommentPos, nextPos)
			// 	if c.Slash > *v.lastCommentPos && c.Slash < nextPos {
			// 		v.buf.WriteString(c.Text())
			// 	}
			// }
			v.buf.WriteString("\n\n") // Markdown newline at the end of every comment group
		}
	}
}

func (v *BlogVisitor) Output(filename string) {
	markdown := blackfriday.Run(v.buf.Bytes())

	err := os.WriteFile(filename, []byte(htmlHeader), fs.ModePerm)
	if err != nil {
		panic(err)
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		panic(err)
	}

	_, err = file.Write(markdown)
	if err != nil {
		panic(err)
	}

	_, err = file.Write([]byte(htmlFooter))
	if err != nil {
		panic(err)
	}
}
