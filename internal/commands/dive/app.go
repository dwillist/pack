package dive

import (
	"fmt"
	"github.com/buildpacks/lifecycle"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/runtime/ui"
	"github.com/wagoodman/dive/runtime/ui/key"
	"github.com/wagoodman/dive/runtime/ui/view"
	"github.com/wagoodman/dive/runtime/ui/viewmodel"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/wagoodman/dive/runtime/ui/layout"

	"github.com/buildpacks/pack"
)

type App struct {
	gui         *gocui.Gui
	controllers *ui.Controller
	layout      *layout.Manager
}

func (a *App) Run() error {
	go func() {
		time.Sleep(1 * time.Minute)
		a.Quit()
	}()

	if err := a.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		logrus.Error("main loop error: ", err)
		return err
	}
	return nil
}

func (a *App) Quit() error {

	// profileObj.Stop()
	// onExit()

	return gocui.ErrQuit
}

type AppOptions struct {
	DiveResult *pack.DiveResult
	GUI        *gocui.Gui
	Debug		bool
}

type CNBDetailsModel struct {
	Lookup pack.LayerLookup
	BuildMetadata lifecycle.BuildMetadata
}

type cnbDiveModels struct {
	filetree *viewmodel.FileTree
	compareMode *viewmodel.LayerCompareMode
	layerSelection *viewmodel.LayerSelection
	layerSetState *viewmodel.LayerSetState
	// TODO: move me to the models folder
	details *CNBDetailsModel
}

func initializeModels(gui *gocui.Gui, diveResult *pack.DiveResult) (cnbDiveModels, error) {
	trees := diveResult.CNBImage.Trees
	layers := diveResult.CNBImage.Layers

	// TODO: do we need to check this??
	firstTree := trees[0]
	firstLayer := layers[0]

	// fileTreeModel initialization
	fileTreeModel, err := viewmodel.NewFileTreeViewModel(firstTree, trees, diveResult.TreeStack)
	if err != nil {
		return cnbDiveModels{}, err
	}

	// compareMode
	var compareMode viewmodel.LayerCompareMode
	switch mode := viper.GetBool("layer.show-aggregated-changes"); mode {
	case true:
		compareMode = viewmodel.CompareAllLayers
	case false:
		compareMode = viewmodel.CompareSingleLayer
	default:
		return cnbDiveModels{}, fmt.Errorf("unknown layer.show-aggregated-changes value: %v", mode)
	}

	layerSelection := viewmodel.LayerSelection{
		Layer:           firstLayer,
		BottomTreeStart: 0,
		BottomTreeStop:  0,
		TopTreeStart:    0,
		TopTreeStop:     0,
	}

	layerSetState := viewmodel.NewLayerSetState(layers, compareMode)
	details := CNBDetailsModel{
		Lookup: diveResult.LayerLookupInfo,
		BuildMetadata: diveResult.BuildMetadata,
	}
	return cnbDiveModels{
		filetree: fileTreeModel,
		compareMode: &compareMode,
		layerSelection: &layerSelection,
		layerSetState: layerSetState,
		details: &details,
	}, nil
}

func initializeViews(gui *gocui.Gui, m cnbDiveModels) (*Views, error) {
	layerView, err := view.NewLayerView(gui, m.layerSetState)
	if err != nil {
		return nil, err
	}

	fileTreeView, err := view.NewFileTreeView(gui, m.filetree)

	statusView := view.NewStatusView(gui)
	statusView.SetCurrentView(layerView)

	filterView := view.NewFilterView(gui)

	detailsView := NewCNBDetailsView(gui, m.details.Lookup, m.details.BuildMetadata)

	return &Views{
		Tree: fileTreeView,
		Layer: layerView,
		Status: statusView,
		Filter: filterView,
		Details: detailsView,
	}, nil
}

// TODO: this looks a bt odd, way too much is private/ specific to a individual controller
// Can we build some meaningfull abstractions here??
func initializeController(g *gocui.Gui, views *view.Views) (*ui.Controller, error) {
	controller := &ui.Controller{
		Gui:   g,
		Views: views,
	}

	views.Layer.AddLayerChangeListener(controller.OnLayerChange)

	// update the status pane when a filetree option is changed by the user
	views.Tree.AddViewOptionChangeListener(controller.OnFileTreeViewOptionChange)

	// update the tree view while the user types into the filter view
	views.Filter.AddFilterEditListener(controller.OnFilterEdit)

	err := controller.OnLayerChange(viewmodel.LayerSelection{
		Layer:           views.Layer.CurrentLayer(),
		BottomTreeStart: 0,
		BottomTreeStop:  0,
		TopTreeStart:    0,
		TopTreeStop:     0,
	})

	if err != nil {
		return nil, err
	}

	return controller, nil
}

type app struct {
	gui *gocui.Gui
	controllers *Controller
	layout    *layout.Manager
}

func (a *app) quit() error {

	// profileObj.Stop()
	// onExit()

	return gocui.ErrQuit
}


func newApp(options AppOptions) (*app, error) {
	var err error
	var myAppSingleton *app
	once.Do(func() {
		var globalHelpKeys []*key.Binding
		// create models
		m, err := initializeModels(options.GUI, options.DiveResult)
		if err != nil {
			return
		}

		// create views
		v, err := initializeViews(options.GUI,m)
		if err != nil {
			return
		}

		controller, err := NewCollection(options.GUI, v)
		if err != nil {
			return
		}
		// controller setup (this should be specific to our implementation)


		lm := layout.NewManager()
		lm.Add(v.Status, layout.LocationFooter)
		lm.Add(v.Filter, layout.LocationFooter)
		lm.Add(NewLayerDetailsCompoundLayout(v.Layer, v.Details), layout.LocationColumn)
		lm.Add(v.Tree, layout.LocationColumn)

		//if options.Debug {
		//	lm.Add(controller.Views.Debug, layout.LocationColumn)
		//}
		options.GUI.Cursor = false
		options.GUI.SetManagerFunc(lm.Layout)

		myAppSingleton = &app{
			gui:         options.GUI,
			controllers: controller,
			layout:      lm,
		}

		var infos = []key.BindingInfo{
			{
				ConfigKeys: []string{"keybinding.quit"},
				OnAction:   myAppSingleton.quit,
				Display:    "Quit",
			},
			{
				ConfigKeys: []string{"keybinding.toggle-view"},
				OnAction:   controller.ToggleView,
				Display:    "Switch view",
			},
			{
				ConfigKeys: []string{"keybinding.filter-files"},
				OnAction:   controller.ToggleFilterView,
				IsSelected: controller.Views.Filter.IsVisible,
				Display:    "Filter",
			},
		}

		globalHelpKeys, err = key.GenerateBindings(options.GUI, "", infos)
		if err != nil {
			return
		}

		controller.Views.Status.AddHelpKeys(globalHelpKeys...)

		// perform the first update and render now that all resources have been loaded
		err = controller.UpdateAndRender()
		if err != nil {
			return
		}
	})

	return myAppSingleton, err
}



//func NewApp(appOptions AppOptions) (*App, error) {
//	var err error
//	once.Do(func() {
//		var controller *Controller
//		//var globalHelpKeys []*key.Binding
//
//		controller, err = NewController(appOptions.GUI, appOptions.DiveResult)
//		if err != nil {
//			return
//		}
//
//		// note: order matters when adding elements to the layout
//		lm := layout.NewManager()
//		lm.Add(controller.views.Status, layout.LocationFooter)
//		lm.Add(NewLayerDetailsCompoundLayout(controller.views.Layer, controller.views.Details), layout.LocationColumn)
//		lm.Add(controller.views.Tree, layout.LocationColumn)
//
//		appOptions.GUI.Cursor = false
//		//g.Mouse = true
//		appOptions.GUI.SetManagerFunc(lm.Layout)
//
//		// var profileObj = profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
//		//
//		// onExit = func() {
//		// 	profileObj.Stop()
//		// }
//
//		appSingleton = &App{
//			gui:         appOptions.GUI,
//			controllers: controller,
//			layout:      lm,
//		}
//
//		// need to set up these keybindings, there is just no preceeding configuration.
//		var infos = []key.BindingInfo{
//			{
//				Key: gocui.KeyCtrlC,
//				//ConfigKeys: []string{"ctrl+c"},
//				OnAction: appSingleton.Quit,
//				Display:  "Quit (ctrl+c)",
//			},
//
//			{
//				Key: gocui.KeyTab,
//				//ConfigKeys: []string{"tab"},
//				OnAction: controller.ToggleView,
//				Display:  "Switch view (tab)",
//			},
//			//{
//			//	ConfigKeys: []string{"keybinding.filter-files"},
//			//	OnAction:   controller.ToggleFilterView,
//			//	IsSelected: controller.views.Filter.IsVisible,
//			//	Display:    "Filter",
//			//},
//		}
//
//		globalHelpKeys, err := key.GenerateBindings(appOptions.GUI, "", infos)
//		if err != nil {
//			logrus.Error(globalHelpKeys)
//			return
//		}
//
//		controller.views.Status.AddHelpKeys(globalHelpKeys...)
//
//		// perform the first update and render now that all resources have been loaded
//		err = controller.UpdateAndRender()
//		if err != nil {
//			return
//		}
//
//	})
//
//	return appSingleton, err
//}
