package main

import (
	"database/sql"
	"image"
	"log"

	"github.com/mjl-/duit"
)

type connUI struct {
	cc          configConnection
	db          *sql.DB
	databaseBox *duit.Box // for the selected database, after connecting

	unconnected duit.UI

	*duit.Box
}

func (ui *connUI) layout() {
	dui.MarkLayout(ui)
}

func newConnUI(cc configConnection) (ui *connUI) {
	var databaseList *duit.List

	var noDBUI duit.UI = duit.NewMiddle(&duit.Label{Text: "select a database on the left"})

	var connecting, databases duit.UI

	cancel := &duit.Button{
		Text: "cancel",
		Click: func(r *duit.Event) {
			log.Printf("todo: should cancel new connection\n")
		},
	}

	ui = &connUI{cc: cc}
	connect := &duit.Button{
		Text:     "connect",
		Colorset: &dui.Primary,
		Click: func(result *duit.Event) {
			defer ui.layout()
			ui.Box.Kids = duit.NewKids(connecting)

			go func() {
				setStatus := func(err error) {
					dui.Call <- func() {
						defer dui.MarkLayout(ui.databaseBox)
						ui.databaseBox.Kids = duit.NewKids(&duit.Label{Text: "error: " + err.Error()})
					}
				}
				db, err := sql.Open("postgres", cc.connectionString(cc.Database))
				if err != nil {
					setStatus(err)
					return
				}
				q := `select coalesce(json_agg(datname order by datname asc), '[]') from pg_database where not datistemplate`
				var dbNames []string
				err = parseRow(db.QueryRow(q), &dbNames, "parsing list of databases")
				if err != nil {
					setStatus(err)
					db.Close()
					return
				}

				dbValues := make([]*duit.ListValue, len(dbNames))
				var sel *duit.ListValue
				for i, name := range dbNames {
					lv := &duit.ListValue{
						Text:     name,
						Value:    newDBUI(ui, name),
						Selected: name == cc.Database,
					}
					dbValues[i] = lv
					if lv.Selected {
						sel = lv
						sel.Value.(*dbUI).init()
					}
				}

				dui.Call <- func() {
					defer ui.layout()
					ui.db = db
					disconnect.Disabled = false
					databaseList.Values = dbValues
					nui := noDBUI
					if sel != nil {
						nui = sel.Value.(*dbUI)
					}
					ui.Box.Kids = duit.NewKids(databases)
					ui.databaseBox.Kids = duit.NewKids(nui)
				}
			}()
		},
	}
	edit := &duit.Button{
		Text: "edit",
		Click: func(r *duit.Event) {
			defer ui.layout()
			ui.Box.Kids = duit.NewKids(newSettingsUI(ui.cc, func() {
				defer ui.layout()
				ui.Box.Kids = duit.NewKids(ui.unconnected)
			}))
		},
	}
	ui.unconnected = duit.NewMiddle(
		&duit.Box{
			Margin: image.Pt(4, 2),
			Kids:   duit.NewKids(connect, edit),
		},
	)
	connecting = duit.NewMiddle(
		&duit.Box{
			Margin: image.Pt(4, 2),
			Kids:   duit.NewKids(&duit.Label{Text: "connecting..."}, cancel),
		},
	)
	databaseList = &duit.List{
		Changed: func(index int, result *duit.Event) {
			defer dui.MarkLayout(ui.databaseBox)
			lv := databaseList.Values[index]
			nui := noDBUI
			shouldConnect := false
			if lv.Selected {
				dbUI := lv.Value.(*dbUI)
				nui = dbUI
				shouldConnect = dbUI.db == nil
			}
			ui.databaseBox.Kids = duit.NewKids(nui)
			if shouldConnect {
				go lv.Value.(*dbUI).init()
			}
		},
	}
	ui.databaseBox = &duit.Box{
		Kids: duit.NewKids(noDBUI),
	}
	databases = &duit.Horizontal{
		Split: func(width int) []int {
			if hideLeftBars {
				return []int{0, width}
			}
			first := dui.Scale(200)
			if first > width/2 {
				first = width / 2
			}
			return []int{first, width - first}
		},
		Kids: duit.NewKids(
			&duit.Box{
				Kids: duit.NewKids(
					duit.CenterUI(duit.SpaceXY(4, 2), &duit.Label{Text: "databases", Font: bold}),
					duit.NewScroll(databaseList),
				),
			},
			ui.databaseBox,
		),
	}
	ui.Box = &duit.Box{
		Kids: duit.NewKids(ui.unconnected),
	}
	return
}

func (ui *connUI) disconnect() {
	// xxx todo: close all lower dbUI db connections
	ui.db.Close()
	ui.db = nil
	ui.Box.Kids = duit.NewKids(ui.unconnected)
	dui.MarkLayout(ui)
}
