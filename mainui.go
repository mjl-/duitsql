package main

import (
	"image"
	"sort"

	"github.com/mjl-/duit"
	"github.com/mjl-/filterlist"
)

type mainUI struct {
	hideLeftBars   bool
	noConnectionUI duit.UI
	disconnect     *duit.Button
	status         *duit.Label
	connections    *filterlist.Filterlist
	connectionBox  *duit.Box
	splitKid       *duit.Kid
	split          *duit.Split
	duit.Box
}

func newMainUI(configs []connectionConfig) (ui *mainUI) {
	ui = &mainUI{}

	ui.connectionBox = &duit.Box{}
	ui.noConnectionUI = middle(label("select a connection on the left"))
	ui.connectionBox.Kids = duit.NewKids(ui.noConnectionUI)

	connectionValues := make([]*duit.ListValue, len(configs)+1)
	for i, c := range configs {
		lv := &duit.ListValue{
			Text:  c.Name,
			Value: newConnUI(c),
		}
		connectionValues[i] = lv
	}
	connectionValues[len(connectionValues)-1] = &duit.ListValue{
		Text:  "<new>",
		Value: nil, // indicates this is the "new" entry
	}
	ui.sortConnections(connectionValues)

	ui.connections = filterlist.NewFilterlist(dui, &duit.List{Values: connectionValues})
	ui.connections.List.Changed = func(index int) (e duit.Event) {
		ui.disconnect.Disabled = true
		dui.MarkDraw(ui.disconnect)
		lv := ui.connections.List.Values[index]
		if !lv.Selected {
			ui.connectionBox.Kids = duit.NewKids(ui.noConnectionUI)
			dui.MarkLayout(ui)
			return
		}
		if lv.Value == nil {
			nop := func() {}
			sUI := newSettingsUI(connectionConfig{Type: "postgres"}, true, nop)
			ui.connectionBox.Kids = duit.NewKids(sUI)
			dui.MarkLayout(ui)
			dui.Focus(sUI.name)
			return
		}
		cUI := lv.Value.(*connUI)
		ui.disconnect.Disabled = cUI.db == nil
		ui.connectionBox.Kids = duit.NewKids(cUI)
		dui.MarkLayout(ui)
		if cUI.db == nil {
			dui.Focus(cUI.connect)
		}
		return
	}

	toggleSlim := &duit.Button{
		Text: "toggle left",
		Click: func() (e duit.Event) {
			ui.hideLeftBars = !ui.hideLeftBars
			ui.ensureLeftBars()
			dui.MarkLayout(nil)
			return
		},
	}
	ui.disconnect = &duit.Button{
		Text:     "disconnect",
		Disabled: true,
		Click: func() (e duit.Event) {
			dui.MarkLayout(nil)
			l := ui.connections.List.Selected()
			if len(l) != 1 {
				return
			}
			lv := ui.connections.List.Values[l[0]]
			cUI := lv.Value.(*connUI)
			cUI.disconnect()
			dui.Focus(ui.connections.Search)
			return
		},
	}
	ui.status = &duit.Label{}

	ui.split = &duit.Split{
		Gutter:     1,
		Background: dui.Gutter,
		Split: func(width int) []int {
			return ui.splitDimensions(width)
		},
		Kids: duit.NewKids(
			&duit.Box{
				Kids: duit.NewKids(
					duit.CenterUI(duit.SpaceXY(4, 2), &duit.Label{Text: "connections", Font: bold}),
					ui.connections,
				),
			},
			ui.connectionBox,
		),
	}
	ui.Box.Kids = duit.NewKids(
		&duit.Box{
			Padding: duit.SpaceXY(4, 2),
			Margin:  image.Pt(4, 2),
			Kids:    duit.NewKids(toggleSlim, ui.disconnect, ui.status),
		},
		ui.split,
	)
	ui.splitKid = ui.Box.Kids[1]
	ui.splitKid.ID = "connections"
	return
}

func (ui *mainUI) splitDimensions(width int) []int {
	if ui.hideLeftBars {
		return []int{0, width}
	}
	first := dui.Scale(125)
	if first > width/2 {
		first = width / 2
	}
	return []int{first, width - first}
}

func (ui *mainUI) ensureLeftBars() {
	width := ui.splitKid.R.Dx()
	ui.split.Dimensions(dui, ui.splitDimensions(width))
	if cUI, ok := ui.connectionBox.Kids[0].UI.(*connUI); ok {
		cUI.ensureLeftBars()
	}
}

func (ui *mainUI) layout() {
	dui.MarkLayout(ui)
}

// essentially opens the "new settings" ui, but with the given config filled in.
func (ui *mainUI) duplicateSettings(c connectionConfig) {
	var sUI *settingsUI
	for _, lv := range ui.connections.List.Values {
		if lv.Value == nil {
			lv.Selected = true
			nop := func() {}
			sUI = newSettingsUI(c, true, nop)
			ui.connectionBox.Kids = duit.NewKids(sUI)

			if !ui.connections.Match(c.Name) {
				ui.connections.Search.Text = lv.Text
				ui.connections.Search.Changed("")
			}
		} else {
			lv.Selected = false
		}
	}
	ui.layout()
	dui.Focus(sUI.name)
}

// sortConnections sorts  values for a connection list by label (config name).
// "<new>" is always at the end.
func (ui *mainUI) sortConnections(l []*duit.ListValue) {
	sort.Slice(l, func(i, j int) bool {
		a, b := l[i], l[j]
		if a.Value == nil || b.Value == nil {
			return a.Value != nil
		}
		return a.Text < b.Text
	})
}

// adds config as new connection to connectionst list.
// Called from main loop.
func (ui *mainUI) addNewConnection(c connectionConfig) {
	ui.connections.List.Unselect(nil)
	cUI := newConnUI(c)
	lv := &duit.ListValue{
		Text:     c.Name,
		Value:    cUI,
		Selected: true,
	}
	ui.connections.Values = append(ui.connections.Values, lv)
	ui.sortConnections(ui.connections.Values)
	if !ui.connections.Match(c.Name) {
		ui.connections.Search.Text = c.Name
	}
	ui.connections.Filter()
	ui.connectionBox.Kids = duit.NewKids(cUI)
	ui.layout()
}

// updates the selected connection with the new config.
// called from main loop
func (ui *mainUI) updateSelectedConnection(c connectionConfig) {
	index := ui.connections.List.Selected()[0]
	lv := ui.connections.List.Values[index]
	lv.Value.(*connUI).config = c
	lv.Text = c.Name
	ui.sortConnections(ui.connections.List.Values)
	ui.sortConnections(ui.connections.Values)
	if !ui.connections.Match(lv.Text) {
		ui.connections.Search.Text = lv.Text
		ui.connections.Search.Changed(lv.Text)
	}
	dui.MarkDraw(ui)
}

// deletes the selected connection.
// called from main loop
func (ui *mainUI) deleteSelectedConnection() {
	values := []*duit.ListValue{}
	for _, lv := range ui.connections.Values {
		if !lv.Selected {
			values = append(values, lv)
		}
	}
	ui.connections.Values = values
	ui.connections.Filter()
	ui.connectionBox.Kids = duit.NewKids(ui.noConnectionUI)
	ui.layout()
}

// saveConnects saves all current connection configs..
func (ui *mainUI) saveConnections() {
	l := []connectionConfig{}
	for _, lv := range ui.connections.Values {
		if lv.Value != nil {
			l = append(l, lv.Value.(*connUI).config)
		}
	}

	go saveConnectionConfigs(l)
}
