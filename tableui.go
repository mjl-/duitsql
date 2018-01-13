package main

import (
	"github.com/mjl-/duit"
)

type tableUI struct {
	dbUI     *dbUI
	name     string
	resultUI *resultUI
	tabsUI   *duit.Tabs
	duit.Box
}

func newTableUI(dbUI *dbUI, name string) *tableUI {
	ui := &tableUI{
		dbUI: dbUI,
		name: name,
	}
	return ui
}

func (ui *tableUI) init() {
	if ui.resultUI != nil {
		return
	}
	query := `select * from ` + ui.name
	ui.resultUI = newResultUI(ui.dbUI, query)
	tsUI := newTableStructUI(ui.dbUI, ui.name)
	tsUI.init()
	ui.tabsUI = &duit.Tabs{
		Buttongroup: &duit.Buttongroup{
			Texts: []string{
				"Data",
				"Structure",
			},
		},
		UIs: []duit.UI{
			ui.resultUI,
			tsUI,
		},
	}
	ui.Box.Kids = duit.NewKids(ui.tabsUI)
	dui.MarkLayout(nil) // xxx
	go ui.resultUI.load()
}
