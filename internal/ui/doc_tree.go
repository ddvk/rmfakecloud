package ui

import (
	"github.com/ddvk/rmfakecloud/internal/ui/methods"
	"github.com/ddvk/rmfakecloud/internal/ui/templates"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
)

// documentTreeFromBlob builds the web UI tree from the hash-tree (sync 1.5+ storage),
// merging builtin templates/methods. Used for both sync backends so types and previews match.
func documentTreeFromBlob(blobHandler blobHandler, uid string) (*viewmodel.DocumentTree, error) {
	hashTree, err := blobHandler.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}
	tree := viewmodel.DocTreeFromHashTree(hashTree)
	tDir := templates.BuiltinTemplatesDirectory()
	if len(tree.Templates) > 0 {
		if d, ok := tree.Templates[0].(*viewmodel.Directory); ok {
			tDir.Entries = append(tDir.Entries, d.Entries...)
		}
	}
	tree.Templates = []viewmodel.Entry{tDir}
	mDir := methods.BuiltinMethodsDirectory()
	if len(tree.Methods) > 0 {
		if d, ok := tree.Methods[0].(*viewmodel.Directory); ok {
			mDir.Entries = append(mDir.Entries, d.Entries...)
		}
	}
	tree.Methods = []viewmodel.Entry{mDir}
	for _, e := range tDir.Entries {
		if doc, ok := e.(*viewmodel.Document); ok {
			if o, err := blobHandler.GetDocumentOrientation(uid, doc.ID); err == nil {
				doc.Orientation = o
			}
		}
	}
	for _, e := range mDir.Entries {
		if doc, ok := e.(*viewmodel.Document); ok {
			if o, err := blobHandler.GetDocumentOrientation(uid, doc.ID); err == nil {
				doc.Orientation = o
			}
		}
	}
	return tree, nil
}
