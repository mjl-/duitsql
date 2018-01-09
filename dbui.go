package main

import (
	"database/sql"

	"github.com/mjl-/duit"
)

type dbUI struct {
	connUI *connUI
	dbName string
	db     *sql.DB

	tableList *duit.List
	viewUI    *duit.Box // holds 1 kid, the editUI, tableUI or placeholder label

	*duit.Box
}

func newDBUI(cUI *connUI, dbName string) (ui *dbUI) {
	ui = &dbUI{
		connUI: cUI,
		dbName: dbName,
	}
	busy := duit.NewMiddle(&duit.Label{Text: "listing tables..."})

	ui.Box = &duit.Box{
		Kids: duit.NewKids(busy),
	}
	return
}

func (ui *dbUI) layout() {
	dui.MarkLayout(ui)
}

func (ui *dbUI) init() {
	setStatus := func(err error) {
		dui.Call <- func() {
			defer ui.layout()
			ui.Box.Kids = duit.NewKids(&duit.Label{Text: "error: " + err.Error()})
		}
	}

	db, err := sql.Open("postgres", ui.connUI.cc.connectionString(ui.dbName))
	if err != nil {
		setStatus(err)
		return
	}
	q := `
		select coalesce(json_agg(x.name order by internal asc, name asc), '[]') from (
			select table_schema in ('pg_catalog', 'information_schema') as internal, table_schema || '.' || table_name as name
			from information_schema.tables
		) x
	`
	var tables []string
	err = parseRow(db.QueryRow(q), &tables, "listing tables in database")
	if err != nil {
		setStatus(err)
		return
	}

	editUI := newEditUI(ui)
	values := make([]*duit.ListValue, 1+len(tables))
	values[0] = &duit.ListValue{
		Selected: true,
		Text:     "<sql>",
		Value:    editUI,
	}
	for i, tabName := range tables {
		values[i+1] = &duit.ListValue{
			Text:  tabName,
			Value: newTableUI(ui, "select * from "+tabName),
		}
	}

	dui.Call <- func() {
		defer ui.layout()
		ui.db = db
		ui.tableList = &duit.List{
			Values: values,
			Changed: func(index int, r *duit.Event) {
				defer ui.layout()
				lv := ui.tableList.Values[index]
				var selUI duit.UI
				if !lv.Selected {
					selUI = duit.NewMiddle(&duit.Label{Text: "select <sql>, or a a table or view on the left"})
				} else {
					selUI = lv.Value.(duit.UI)
					tUI, ok := selUI.(*tableUI)
					if ok && tUI.grid == nil {
						go tUI.load()
					}
				}
				ui.viewUI.Kids = duit.NewKids(selUI)
			},
		}
		ui.viewUI = &duit.Box{
			Kids: duit.NewKids(editUI),
		}
		ui.Box.Kids = duit.NewKids(
			&duit.Horizontal{
				Split: func(width int) []int {
					first := dui.Scale(200)
					if first > width/2 {
						first = width / 2
					}
					return []int{first, width - first}
				},
				Kids: duit.NewKids(
					&duit.Box{
						Kids: duit.NewKids(
							duit.CenterUI(duit.SpaceXY(4, 2), &duit.Label{Text: "tables", Font: bold}),
							duit.NewScroll(ui.tableList),
						),
					},
					ui.viewUI,
				),
			},
		)
	}
}
