// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlxWrapper

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/mattermost/mattermost/tools/mattermost-govet/util"
	"golang.org/x/tools/go/analysis"
)

const sqlstorePackagePath = util.ServerModulePath + "/channels/store/sqlstore"

var Analyzer = &analysis.Analyzer{
	Name: "sqlxWrapper",
	Doc:  "check for direct SQL method calls on embedded .DB or .Tx fields that bypass the wrapper's query timeout",
	Run:  run,
}

// sqlMethods is the set of methods that execute SQL and must be called on the
// wrapper (which enforces the query timeout) rather than on the embedded field.
var sqlMethods = map[string]bool{
	"Exec": true, "ExecContext": true,
	"Get": true, "GetContext": true,
	"NamedExec": true, "NamedExecContext": true,
	"NamedQuery": true,
	"Query":      true, "QueryContext": true,
	"QueryRow": true, "QueryRowContext": true,
	"QueryRowx": true, "QueryRowxContext": true,
	"Queryx": true, "QueryxContext": true,
	"Select": true, "SelectContext": true,
}

// wrapperTypes maps the embedded field name to the wrapper struct type that
// owns it.  These are the only types whose embedded fields we want to police.
var wrapperTypes = map[string]string{
	"DB": "sqlxDBWrapper",
	"Tx": "sqlxTxWrapper",
}

func run(pass *analysis.Pass) (interface{}, error) {
	if !strings.HasPrefix(pass.Pkg.Path(), sqlstorePackagePath) {
		return nil, nil
	}

	for _, file := range pass.Files {
		ignoredLines := map[int]bool{}
		for _, cg := range file.Comments {
			for _, c := range cg.List {
				if strings.Contains(c.Text, "nolint:sqlxWrapper") {
					ignoredLines[pass.Fset.Position(c.Pos()).Line] = true
				}
			}
		}

		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Match: <receiver>.<field>.<method>(...)
			// where <receiver>.<field> is, e.g., w.DB or txn.Tx
			outerSel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if !sqlMethods[outerSel.Sel.Name] {
				return true
			}

			innerSel, ok := outerSel.X.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			fieldName := innerSel.Sel.Name
			expectedWrapper, isManagedField := wrapperTypes[fieldName]
			if !isManagedField {
				return true
			}

			// Verify the receiver is actually a *sqlxDBWrapper / *sqlxTxWrapper
			// so we don't accidentally flag unrelated types that happen to have
			// a field named DB or Tx.
			receiverType := pass.TypesInfo.TypeOf(innerSel.X)
			if receiverType == nil {
				return true
			}
			ptr, ok := receiverType.(*types.Pointer)
			if !ok {
				return true
			}
			named, ok := ptr.Elem().(*types.Named)
			if !ok {
				return true
			}
			if named.Obj().Name() != expectedWrapper {
				return true
			}

			if ignoredLines[pass.Fset.Position(outerSel.Sel.Pos()).Line] {
				return true
			}

			pass.Reportf(
				outerSel.Sel.Pos(),
				"direct call to .%s.%s() bypasses the wrapper's query timeout; call .%s() on the wrapper directly",
				fieldName, outerSel.Sel.Name, outerSel.Sel.Name,
			)
			return true
		})
	}
	return nil, nil
}
