// # Literate Programming in Go

// This is an experiment in creating literate go files
package main

/*
I was curious if I could create something that parses go packages and generates blog-style pages representing all of the code, documented via the comments. I'm hoping that this will be useful for others, but I'm pretty sure it'll be useful for me (I feel like I'm often doing little experiments here and there, so it'd be nice to have a place to throw them all). This file here will be my very first experiment.

Let's give it a shot. My high-level strategy will be to do some Abstract-Syntax-Tree (AST) crawling, and then template the data into markdown, then turn that markdown into HTML files in some static-blog-sort-of-way. I ended up finding `github.com/russross/blackfriday/v2` which is a super easy to use markdown to HTML (and other) conversion engine.

The file structure will be 100% compilable Go code, then I'll put markdown directly into comment blocks and render the comments to HTML. I opted to use AST-crawling logic which was way harder than I originally anticipated - but it seems to work. It's probably the most flexible solution, because now while I'm walking the AST, I can setup references from here-to-there and link things together however I please. With a flat mapping (ie read file, parse comments and code, then just spit them to a file), things like this would be much more difficult.
*/

// ## Imports
/*
I'm importing some fairly standard packages, each of which is described below:
*/
import (
	// First we need to import some packages that can do file reading and writing
	"fmt"
	"bytes"
	"io/fs"
	"os"
	"math"

	// These are the golang ast-related packages for AST walking, parsing and printing
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	// blackfriday is used to parse markdown and convirt it to HTML (or other)
	"github.com/russross/blackfriday/v2" // Test comment

	// We will also embed some other template-type files into this binary
	_ "embed"
)

/*
In this next section we define both an html header `header.html` and footer `footer.html` which will be used to generate the final HTML. I need to wrap the inner blackfriday-generated html in something so that when it displays in a browser it looks nice. So in my header file I have a nav bar and some imported CSS. We embed both of these files directly into the browser, this makes it super easy because we don't have to pass many files around (I should probably do it with the CSS too). I think ideally, I also should have a single template.html file and use Go's templating library to build the final HTML, but that's more work - and I'm feeling kind of lazy.
*/

//go:embed header.html
var htmlHeader string
//go:embed footer.html
var htmlFooter string

// Let's have a main function that just wraps some high level logic that we will write below
func main() {
	generatePackage(".")
}

/*
`generatePackage` will be the main workhorse function, essentially reading a directory and generating an HTML output for the entire package.
*/
func generatePackage(dir string) {

	// We essentially parse the directory into the fileset and list of packages
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Then we loop over the packages
	for _, pkg := range packages {
		fmt.Println("Parsing", pkg.Name)
		tokenStart := pkg.Pos()

		// We build a blog visitor (which implements the ast.Visitor interface).
		// This will be used to walk the entire AST!
		bv := &BlogVisitor{
			buf: &bytes.Buffer{},
			pkg: pkg,
			fset: fset,
			lastCommentPos: &tokenStart,
		}

		// We start walking our BlogVisitor `bv` through the AST in a depth-first way.
		ast.Walk(bv, pkg)

		// Because of how I wrote this, it's hard to handle comments after the last AST node.
		// So we will process those here, at the end, once we've walked the whole tree
		if bv.file != nil {
			// We want to process the rest of the comments, so we process until math.MaxInt
			bv.handleComments(token.Pos(math.MaxInt))
		}

		// Finally, we dump the BlogVisitor buffered data into the final HTML file.
		bv.Output("output.html")
	}
}

/*
I couldn't really figure out a good way to combine the formatting of each node type, because everything may need to do something very slightly different.

For formatting function declarations. We just want to remove the header docs (printed elsewhere) and print out the node in a code block.
*/
func formatFunc(buf *bytes.Buffer, fset *token.FileSet, decl ast.FuncDecl, cGroups []*ast.CommentGroup) {
	decl.Doc = nil // nil the Doc field so that we don't print it

	// Build a CommentedNode. This is important, if you don't attach the comment
	// group to the node then the comments inside the function will be removed!
	commentedNode := printer.CommentedNode{
		Node:     &decl,
		Comments: cGroups,
	}
	formatNode(buf, fset, &commentedNode)
}

// For formatting the General tokens: imports, types, variables, or constants things are slightly more complex. For imports and types, we want to remove the docs and print as usual. But on variables and constants, we might have a `go:embed` statement (or something important) above it. This does cause a bug, though: if we don't put a space above the `go:embed` then we will include that as a comment in the code block printout.
func formatGen(buf *bytes.Buffer, fset *token.FileSet, decl ast.GenDecl, cGroups []*ast.CommentGroup) {
	commentedNode := printer.CommentedNode{
		Node:     &decl,
		Comments: cGroups,
	}

	if decl.Tok == token.IMPORT || decl.Tok == token.TYPE{
		decl.Doc = nil // nil the Doc field so that we don't print it
	} else if decl.Tok == token.CONST || decl.Tok == token.VAR {
		// Don't nil the documentation
	}

	formatNode(buf, fset, &commentedNode)
}

// The `formatNode` function is used to do the final printout to the buffer. We basically configure the printer to turn the tabs into 2 spaces. This is mostly for readability on a browser. After that we print the node out
func formatNode(buf *bytes.Buffer, fset *token.FileSet, node any) {
	config := printer.Config{
		Mode:     printer.UseSpaces,
		Tabwidth: 2,
	}
	buf.WriteString("\n```go\n")
	err := config.Fprint(buf, fset, node)
	buf.WriteString("\n```\n")
	if err != nil {
		panic(err)
	}
}

/*
First we define the BlogVisitor struct and all of the data that resides within.
*/
type BlogVisitor struct {
	buf *bytes.Buffer // The buffered blog output
	pkg *ast.Package  // The package that we are processing
	fset *token.FileSet // The fileset of the package we are processing
	file *ast.File // The file we are currently processing (Can be nil if we haven't started processing a file yet!)
	cmap ast.CommentMap // The comment map of the file we are processing

	lastCommentPos *token.Pos // The token.Pos of the last comment or node that we processed
}

/*
Next we implment the `ast.Visitor` interface by implementing this function. Notably, the input is the current node we are processing, which is passed in by the `ast.Walk` function. This value can be null (I assume at the leafs of the AST). We decide what to return back to the `ast.Walk` function: either a visitor (if we want to keep going), or `nil` (if we want to stop). This means that we can modify our visitor along the way and pass it throughout the AST.
*/
func (v *BlogVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil { return nil }

	// If we are a package, then just keep searching
	_, ok := node.(*ast.Package)
	if ok { return v }

	// If we are a file, then store some data in the visitor so we can use it later
	file, ok := node.(*ast.File)
	if ok {
		v.file = file
		v.cmap = ast.NewCommentMap(v.fset, file, file.Comments)
		return v
	}

	// If we are a function, do the function formatting
	f, ok := node.(*ast.FuncDecl)
	if ok {
		// Handle comments
		v.handleComments(f.Pos())
		*v.lastCommentPos = node.End()

		// Handle function case
		cgroups := v.cmap.Filter(f).Comments()
		formatFunc(v.buf, v.fset, *f, cgroups)

		return nil
	}

	// If we are a general declaration, do the general declaration formatting
	gen, ok := node.(*ast.GenDecl)
	if ok {
		// Handle comments
		v.handleComments(gen.Pos())
		*v.lastCommentPos = node.End()

		// Handle node
		cgroups := v.cmap.Filter(gen).Comments()
		formatGen(v.buf, v.fset, *gen, cgroups)

		return nil
	}

	// If all else fails, then keep looking
	return v
}

/*
The way that the AST is parsed is that CommentGroups are separated from the AST itself, this means that we need to print them another way. It seems pretty stable to just maintain a pointer `v.lastCommentPos` to the location of the last comment or node that we printed, then print all of the new comments that are between that point and the next token that we are about to process. That kind of looks like this:

```
[Code]    <- v.lastCommentPos

[Comment] } - Print this one
[Comment] } - Print this one

[Code]   <- nextPos
```

Comments are blocks of text, so we render them as paragraphs, and not code blocks
*/
func (v *BlogVisitor) handleComments(nextPos token.Pos) {
	// Try to printout any comments that are next in line
	if v.file != nil {
		for _, cgroup := range v.file.Comments {
			if cgroup.Pos() > *v.lastCommentPos && cgroup.Pos() < nextPos {
				// Note: the .Text() function removes all the comment (`//` and `/* */`) markers for us!
				v.buf.WriteString(cgroup.Text())
			}

			// Markdown newline at the end of every comment group
			v.buf.WriteString("\n\n")
		}
	}
}

/*
The final challenge is to print the data to a file. So we open a file, and write the following in order:

1. header.html
2. HTML Generated Markdown from AST
3. footer.html
*/
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

	err = file.Close()
	if err != nil {
		panic(err)
	}
}

/*
## Conclusions and future work

I like the general idea of what I've created, although it currently feels very hacked together. It was a good learning experience for me on how AST parsing/walking works. I feel like this will be useful for situations where someone wants to maintain a code-focused blog but they don't want to constantly embed and maintain code blurbs into their markdown. The nice part here is that all you do is code your program, then you can comment it into a blog post. Then whenever you want to change your code, you can just regenerate your blog post. No more having code in two places!

There's obviously tons of room for improvement for this:

1. Better templating of HTML (and theming)
2. Grouping multiple packages into a single frontend page
3. Handling multiple files

There's also a few things that I think would be cool additions:

1. Referencing code blocks (or maybe even other packages in other pages?) via hyperlinks
2. Document layout is basically impossible to control. It'd be cool to reference (maybe minimized) code blocks in other sections. It'd also be cool to have the ability to hide them in the original section
3. It'd be cool to have some sort of "runnable" instance of the code (maybe via webassembly?)
4. Git processing to display a git tree, and pull article date updates

I think in the future I might add some kind of `//lit:command` syntax to solve some of these problems, but for now - I'm pretty happy with it.
*/
