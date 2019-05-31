package filterlist

import (
	"strings"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
)

// Filtergridlist is the full-height containerUI holding the search field and a scrollable gridlist.
type Filtergridlist struct {
	duit.Box

	Search   *duit.Field     // Search box, do not set .Keys or .Changed.
	Rows     []*duit.Gridrow // All possible values for the gridlist, as passed in initially. Gridlist.Rows will differ after searching.
	Gridlist *duit.Gridlist  // Display currently matching values. Gridlist.Changed will be called when the selection changes.
	scroll   *duit.Scroll
	dui      *duit.DUI
}

var _ duit.UI = &Filtergridlist{}

// NewFiltergridlist creates a Filtergridlist.
// Initial values from the gridlist are remembered as all possible values.
// These are filterable through the search field, by substring match. Ctrl-f triggers completion with prefix-match.
func NewFiltergridlist(dui *duit.DUI, gl *duit.Gridlist) (ui *Filtergridlist) {
	ui = &Filtergridlist{
		Gridlist: gl,
		dui:      dui,
	}
	ui.Search = &duit.Field{
		Placeholder: "search...",
		Keys: func(k rune, m draw.Mouse) (e duit.Event) {
			switch k {
			case '\n':
				sel := ui.Gridlist.Selected()
				for _, row := range ui.Rows {
					row.Selected = false
				}
				index := -1
				if len(ui.Gridlist.Rows) > 0 {
					row := ui.Gridlist.Rows[0]
					row.Selected = true
					index = 0
				}
				if ui.Gridlist.Changed != nil && (len(sel) != 1 || sel[0] != 0) {
					e = ui.Gridlist.Changed(index)
				}
				dui.MarkDraw(ui.Gridlist)
				e.Consumed = true

			case 'f' & 0x1f:
				// Completion. The gridlist only contains values that has a substring match.
				e.Consumed = true
				var s string
			Outer:
				for _, row := range ui.Gridlist.Rows {
					for _, v := range row.Values {
						if !strings.HasPrefix(v, ui.Search.Text) {
							continue
						}
						if s == "" {
							s = v
							continue
						}
						for i, c := range []byte(v) {
							if i >= len(s) || s[i] != c {
								s = s[:i]
								if s == "" {
									break Outer
								}
								break
							}
						}
					}
				}
				ui.Search.Text = s
				ui.Search.Cursor1 = 0
				ui.Filter()
				e.NeedDraw = true
			}
			return
		},
		Changed: func(_ string) (e duit.Event) {
			ui.Filter()
			return
		},
	}
	ui.Rows = gl.Rows
	ui.scroll = duit.NewScroll(ui.Gridlist)
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
func (ui *Filtergridlist) Match(row *duit.Gridrow) bool {
	for _, s := range row.Values {
		if strings.Contains(s, ui.Search.Text) {
			return true
		}
	}
	return false
}

// Filter refreshes the listed values by filtering against Search.Text.
func (ui *Filtergridlist) Filter() {
	l := []*duit.Gridrow{}
	for _, row := range ui.Rows {
		if ui.Match(row) {
			l = append(l, row)
		}
	}
	ui.Gridlist.Rows = l
	ui.dui.MarkLayout(ui)
}
