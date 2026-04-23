package lintrequire

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// requiredField is an ObjectFact attached to struct fields that have a
// "// required: <reason>" doc comment. Facts propagate across packages.
type requiredField struct {
	Reason string
}

func (*requiredField) AFact()         {}
func (f *requiredField) String() string { return "required: " + f.Reason }

// Analyzer checks that struct fields annotated with // required: are present
// in all composite literals of that struct type.
var Analyzer = &analysis.Analyzer{
	Name:      "lintrequire",
	Doc:       "check that struct fields marked with // required: are present in composite literals",
	Requires:  []*analysis.Analyzer{inspect.Analyzer},
	FactTypes: []analysis.Fact{(*requiredField)(nil)},
	Run:       run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Pass 1: Export facts for struct fields with // required: comments.
	insp.Preorder([]ast.Node{(*ast.StructType)(nil)}, func(n ast.Node) {
		st := n.(*ast.StructType)
		for _, field := range st.Fields.List {
			reason := parseRequiredComment(field.Doc)
			if reason == "" {
				continue
			}
			for _, name := range field.Names {
				obj := pass.TypesInfo.Defs[name]
				if obj != nil {
					pass.ExportObjectFact(obj, &requiredField{Reason: reason})
				}
			}
		}
	})

	// Pass 2: Check composite literals for missing required fields.
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		cl := n.(*ast.CompositeLit)

		structType := resolveStructType(pass, cl)
		if structType == nil {
			return
		}

		// If the literal uses positional syntax (no keys), all fields up to
		// len(cl.Elts) are provided. We only check keyed literals.
		if len(cl.Elts) > 0 {
			if _, isKV := cl.Elts[0].(*ast.KeyValueExpr); !isKV {
				return // positional literal
			}
		}

		// Build set of field names present in the literal.
		present := make(map[string]bool)
		for _, elt := range cl.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				if ident, ok := kv.Key.(*ast.Ident); ok {
					present[ident.Name] = true
				}
			}
		}

		// Check each struct field for the required fact.
		for i := 0; i < structType.NumFields(); i++ {
			f := structType.Field(i)
			var rf requiredField
			if pass.ImportObjectFact(f, &rf) && !present[f.Name()] {
				pass.Reportf(cl.Pos(), `missing required field %q: %s`, f.Name(), rf.Reason)
			}
		}
	})

	return nil, nil
}

// parseRequiredComment scans a comment group for a line matching "// required: <reason>".
// Returns the reason string or "" if not found.
func parseRequiredComment(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	for _, comment := range doc.List {
		text := strings.TrimSpace(comment.Text)
		// Strip the leading "//"
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimSpace(text)
		lower := strings.ToLower(text)
		if strings.HasPrefix(lower, "required:") {
			reason := strings.TrimSpace(text[len("required:"):])
			if reason == "" {
				reason = "field is required"
			}
			return reason
		}
	}
	return ""
}

// resolveStructType extracts the *types.Struct from a composite literal's type,
// unwrapping pointers, named types, and aliases.
func resolveStructType(pass *analysis.Pass, cl *ast.CompositeLit) *types.Struct {
	tv, ok := pass.TypesInfo.Types[cl]
	if !ok {
		return nil
	}
	return extractStruct(tv.Type)
}

func extractStruct(t types.Type) *types.Struct {
	switch u := t.(type) {
	case *types.Struct:
		return u
	case *types.Named:
		return extractStruct(u.Underlying())
	case *types.Pointer:
		return extractStruct(u.Elem())
	default:
		return nil
	}
}
