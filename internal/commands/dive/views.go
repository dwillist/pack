package dive

import (
	"github.com/wagoodman/dive/runtime/ui/view"
)

// Swap this out to be an interface
type Views struct {
	Tree    *view.FileTree
	Layer   *view.Layer
	Status  *view.Status
	Filter  *view.Filter
	Details *CNBDetails
	//Debug   *view.Debug
}

func (views *Views) All() []view.Renderer {
	return []view.Renderer{
		views.Tree,
		views.Layer,
		views.Status,
		views.Filter,
		views.Details,
		//views.Debug,
	}
}
