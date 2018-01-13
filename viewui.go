package main

import (
	"github.com/mjl-/duit"
)

type viewUI struct {
	dbUI     *dbUI
	name     string
	resultUI *resultUI
	tabsUI   *duit.Tabs
	*duit.Box
}

func newViewUI(dbUI *dbUI, name string) *viewUI {
	ui := &viewUI{
		dbUI: dbUI,
		name: name,
		Box:  &duit.Box{},
	}
	return ui
}

// called from main loop
func (ui *viewUI) init() {
	if ui.resultUI != nil {
		return
	}
	query := `select * from ` + ui.name
	ui.resultUI = newResultUI(ui.dbUI, query)
	vsUI := newViewStructUI(ui.dbUI, ui.name)
	vsUI.init()
	ui.tabsUI = &duit.Tabs{
		Buttongroup: &duit.Buttongroup{
			Texts: []string{
				"Data",
				"Structure",
			},
		},
		UIs: []duit.UI{
			ui.resultUI,
			vsUI,
		},
	}
	ui.Box.Kids = duit.NewKids(ui.tabsUI)
	dui.MarkLayout(nil) // xxx
	go ui.resultUI.load()
}
