// Package filterlist provides a Filterlist and Filtergridlist duit.UI containing a duit.List/duit.Gridlist and a duit.Field search box that filters the entries in the list.
package filterlist

import (
	"strings"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
)

// Filterlist is the full-height containerUI holding the search field and a scrollable list.
type Filterlist struct {
	duit.Box

	Search *duit.Field       // Search box, do not set .Keys or .Changed.
	Values []*duit.ListValue // All possible values for the list, as passed in initially. List.Values will differ after searching.
	List   *duit.List        // List displaying currently matching values. List.Changed will be called when the selection changes.
	scroll *duit.Scroll
	dui    *duit.DUI
}

var _ duit.UI = &Filterlist{}

// NewFilterlist creates a Filterlist.
// Initial values from the list are remembered as all possible values.
// These are filterable through the search field, by substring match. Ctrl-f triggers completion with prefix-match.
func NewFilterlist(dui *duit.DUI, list *duit.List) (ui *Filterlist) {
	ui = &Filterlist{
		List: list,
		dui:  dui,
	}
	ui.Search = &duit.Field{
		Placeholder: "search...",
		Keys: func(k rune, m draw.Mouse) (e duit.Event) {
			switch k {
			case '\n':
				sel := ui.List.Selected()
				for _, lv := range ui.Values {
					lv.Selected = false
				}
				index := -1
				if len(ui.List.Values) > 0 {
					lv := ui.List.Values[0]
					lv.Selected = true
					index = 0
				}
				if ui.List.Changed != nil && (len(sel) != 1 || sel[0] != 0) {
					e = ui.List.Changed(index)
				}
				dui.MarkDraw(ui.List)
				e.Consumed = true

			case 'f' & 0x1f:
				// Completion. The list only contains values that has a substring match.
				e.Consumed = true
				var s string
				if len(ui.List.Values) == 1 {
					s = ui.List.Values[0].Text
				} else {
					for _, lv := range ui.List.Values {
						if !strings.HasPrefix(lv.Text, ui.Search.Text) {
							continue
						}
						if s == "" {
							s = lv.Text
							continue
						}
						for i, c := range []byte(lv.Text) {
							if i >= len(s) || s[i] != c {
								s = s[:i]
								break
							}
						}
					}
				}
				ui.Search.Text = s
				ui.Search.Cursor1 = 0
				ui.Search.Changed(ui.Search.Text)
				e.NeedDraw = true
			}
			return
		},
		Changed: func(_ string) (e duit.Event) {
			ui.Filter()
			return
		},
	}
	ui.Values = list.Values
	ui.List = list
	ui.scroll = duit.NewScroll(ui.List)
	ui.scroll.Height = -1
	ui.Box = duit.Box{
		Height: -1,
		Kids: duit.NewKids(
			&duit.Box{
				Padding: duit.SpaceXY(6, 4),
				Kids:    duit.NewKids(ui.Search),
			},
			ui.scroll,
		),
	}
	return
}

// Match returns whether s matches the current Search.Text.
func (ui *Filterlist) Match(s string) bool {
	return strings.Contains(s, ui.Search.Text)
}

// Filter refreshes the listed values by filtering against Search.Text.
func (ui *Filterlist) Filter() {
	nl := []*duit.ListValue{}
	for _, lv := range ui.Values {
		if ui.Match(lv.Text) {
			nl = append(nl, lv)
		}
	}
	ui.List.Values = nl
	ui.dui.MarkLayout(ui)
}
