package main

import (
	"image"
	"sort"

	"github.com/mjl-/duit"
)

type mainUI struct {
	hideLeftBars   bool
	noConnectionUI duit.UI
	disconnect     *duit.Button
	status         *duit.Label
	connectionList *duit.List
	connectionBox  *duit.Box
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

	ui.connectionList = &duit.List{
		Values: connectionValues,
		Changed: func(index int) (e duit.Event) {
			ui.disconnect.Disabled = true
			dui.MarkDraw(ui.disconnect)
			lv := ui.connectionList.Values[index]
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
				dui.Render()
				dui.Focus(sUI.name)
				return
			}
			cUI := lv.Value.(*connUI)
			ui.disconnect.Disabled = cUI.db == nil
			ui.connectionBox.Kids = duit.NewKids(cUI)
			dui.MarkLayout(ui)
			if cUI.db == nil {
				dui.Render()
				dui.Focus(cUI.connect)
			}
			return
		},
	}
	ui.sortConnections()

	toggleSlim := &duit.Button{
		Text: "toggle left",
		Click: func() (e duit.Event) {
			ui.hideLeftBars = !ui.hideLeftBars
			dui.MarkLayout(nil)
			return
		},
	}
	ui.disconnect = &duit.Button{
		Text:     "disconnect",
		Disabled: true,
		Click: func() (e duit.Event) {
			dui.MarkLayout(nil)
			l := ui.connectionList.Selected()
			if len(l) != 1 {
				return
			}
			lv := ui.connectionList.Values[l[0]]
			cUI := lv.Value.(*connUI)
			cUI.disconnect()
			dui.Focus(ui.connectionList)
			return
		},
	}
	ui.status = &duit.Label{}

	ui.Box.Kids = duit.NewKids(
		&duit.Box{
			Padding: duit.SpaceXY(4, 2),
			Margin:  image.Pt(4, 2),
			Kids:    duit.NewKids(toggleSlim, ui.disconnect, ui.status),
		},
		&duit.Split{
			Gutter:     1,
			Background: dui.Gutter,
			Split: func(width int) []int {
				if ui.hideLeftBars {
					return []int{0, width}
				}
				first := dui.Scale(125)
				if first > width/2 {
					first = width / 2
				}
				return []int{first, width - first}
			},
			Kids: duit.NewKids(
				&duit.Box{
					Kids: duit.NewKids(
						duit.CenterUI(duit.SpaceXY(4, 2), &duit.Label{Text: "connections", Font: bold}),
						duit.NewScroll(ui.connectionList),
					),
				},
				ui.connectionBox,
			),
		},
	)
	ui.Box.Kids[1].ID = "connections"
	return
}

func (ui *mainUI) layout() {
	dui.MarkLayout(ui)
}

// essentially opens the "new settings" ui, but with the given config filled in.
func (ui *mainUI) duplicateSettings(c connectionConfig) {
	var sUI *settingsUI
	for _, lv := range ui.connectionList.Values {
		if lv.Value == nil {
			lv.Selected = true
			nop := func() {}
			sUI = newSettingsUI(c, true, nop)
			ui.connectionBox.Kids = duit.NewKids(sUI)
		} else {
			lv.Selected = false
		}
	}
	ui.layout()
	dui.Render()
	dui.Focus(sUI.name)
}

// sortConnections sorts the connectionList values by label (which should be  the config name).
// "<new>" is always at the end.
func (ui *mainUI) sortConnections() {
	l := ui.connectionList.Values
	sort.Slice(l, func(i, j int) bool {
		a, b := l[i], l[j]
		if a.Value == nil || b.Value == nil {
			return a.Value != nil
		}
		return a.Text < b.Text
	})
}

// adds config as new connection to connectionList.
// Called from main loop.
func (ui *mainUI) addNewConnection(c connectionConfig) {
	cUI := newConnUI(c)
	lv := &duit.ListValue{
		Text:     c.Name,
		Value:    cUI,
		Selected: true,
	}
	ui.connectionList.Unselect(nil)
	ui.connectionList.Values = append([]*duit.ListValue{lv}, ui.connectionList.Values...)
	ui.sortConnections()
	ui.connectionBox.Kids = duit.NewKids(cUI)
	ui.layout()
}

// updates the selected connection with the new config.
// called from main loop
func (ui *mainUI) updateSelectedConnection(c connectionConfig) {
	index := ui.connectionList.Selected()[0]
	lv := ui.connectionList.Values[index]
	lv.Value.(*connUI).config = c
	lv.Text = c.Name
	ui.sortConnections()
	dui.MarkDraw(ui.connectionList)
}

// deletes the selected connection.
// called from main loop
func (ui *mainUI) deleteSelectedConnection() {
	values := []*duit.ListValue{}
	for _, lv := range ui.connectionList.Values {
		if !lv.Selected {
			values = append(values, lv)
		}
	}
	ui.connectionList.Values = values
	ui.connectionBox.Kids = duit.NewKids(ui.noConnectionUI)
	ui.layout()
}

// saveConnects saves all current configs from connectionList.
func (ui *mainUI) saveConnections() {
	l := []connectionConfig{}
	for _, lv := range ui.connectionList.Values {
		if lv.Value != nil {
			l = append(l, lv.Value.(*connUI).config)
		}
	}

	go saveConnectionConfigs(l)
}
