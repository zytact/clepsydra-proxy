package auth

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestExecutorCallsGoThroughTheGuard(t *testing.T) {
	executorMethods := map[string]bool{"Execute": true, "ExecuteStream": true, "CountTokens": true, "HttpRequest": true, "PrepareRequest": true}
	entries, errRead := os.ReadDir(".")
	if errRead != nil {
		t.Fatal(errRead)
	}
	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || name == "execution_guard.go" {
			continue
		}
		file, errParse := parser.ParseFile(fset, name, nil, 0)
		if errParse != nil {
			t.Fatal(errParse)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !executorMethods[selector.Sel.Name] {
				return true
			}
			t.Errorf("%s calls %s directly; use the guarded helper in execution_guard.go", fset.Position(call.Pos()), selector.Sel.Name)
			return true
		})
	}
}
