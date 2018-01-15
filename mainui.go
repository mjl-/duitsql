package main

import (
	"image"

	"github.com/mjl-/duit"
)

type mainUI struct {
	hideLeftBars   bool
	noConnectionUI duit.UI
	disconnect     *duit.Button
	connectionList *duit.List
	connectionBox  *duit.Box
	duit.Box
}

func newMainUI(configConnections []configConnection) (ui *mainUI) {
	ui = &mainUI{}

	ui.connectionBox = &duit.Box{}
	ui.noConnectionUI = duit.NewMiddle(duit.SpaceXY(10, 10), label("select a connection on the left"))
	ui.connectionBox.Kids = duit.NewKids(ui.noConnectionUI)

	connectionValues := make([]*duit.ListValue, len(configConnections)+1)
	for i, cc := range configConnections {
		lv := &duit.ListValue{Text: cc.Name, Value: newConnUI(cc)}
		connectionValues[i] = lv
	}
	connectionValues[len(connectionValues)-1] = &duit.ListValue{Text: "<new>", Value: nil}

	ui.connectionList = &duit.List{
		Values: connectionValues,
		Changed: func(index int) (e duit.Event) {
			defer dui.MarkLayout(ui.connectionBox)
			ui.disconnect.Disabled = true
			lv := ui.connectionList.Values[index]
			if !lv.Selected {
				ui.connectionBox.Kids = duit.NewKids(ui.noConnectionUI)
				return
			}
			if lv.Value == nil {
				ui.connectionBox.Kids = duit.NewKids(newSettingsUI(configConnection{Type: "postgres"}, true, func() {}))
				return
			}
			cUI := lv.Value.(*connUI)
			ui.disconnect.Disabled = cUI.db == nil
			ui.connectionBox.Kids = duit.NewKids(cUI)
			return
		},
	}

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
			return
		},
	}
	status := &duit.Label{}

	ui.Box.Kids = duit.NewKids(
		&duit.Box{
			Padding: duit.SpaceXY(4, 2),
			Margin:  image.Pt(4, 0),
			Kids:    duit.NewKids(toggleSlim, ui.disconnect, status),
		},
		&duit.Split{
			Gutter: 1,
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

func (ui *mainUI) duplicateSettings(cc configConnection) {
	for _, lv := range ui.connectionList.Values {
		if lv.Value == nil {
			lv.Selected = true
			ui.connectionBox.Kids = duit.NewKids(newSettingsUI(cc, true, func() {}))
		} else {
			lv.Selected = false
		}
	}
	ui.layout()
}

// called from main loop
func (ui *mainUI) newConnection(cc configConnection) {
	cUI := newConnUI(cc)
	lv := &duit.ListValue{
		Text:     cc.Name,
		Value:    cUI,
		Selected: true,
	}
	ui.connectionList.Unselect(nil)
	ui.connectionList.Values = append([]*duit.ListValue{lv}, ui.connectionList.Values...)
	ui.connectionBox.Kids = duit.NewKids(cUI)
	dui.MarkDraw(ui.connectionList)
	dui.MarkLayout(ui.connectionBox)
}

// called from main loop
func (ui *mainUI) updateConnection(cc configConnection) {
	index := ui.connectionList.Selected()[0]
	lv := ui.connectionList.Values[index]
	lv.Value.(*connUI).cc = cc
	lv.Text = cc.Name
}

// called from main loop
func (ui *mainUI) deleteSelectedConnection() {
	l := []configConnection{}
	for _, lv := range ui.connectionList.Values {
		if !lv.Selected && lv.Value != nil {
			l = append(l, lv.Value.(*connUI).cc)
		}
	}
	go func() {
		saveConfigConnections(l)

		dui.Call <- func() {
			nvalues := []*duit.ListValue{}
			for _, lv := range ui.connectionList.Values {
				if !lv.Selected {
					nvalues = append(nvalues, lv)
				}
			}
			ui.connectionList.Values = nvalues
			ui.connectionBox.Kids = duit.NewKids(ui.noConnectionUI)
			dui.MarkLayout(nil)
		}
	}()
}
