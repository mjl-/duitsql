package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/mjl-/duit"
)

type dbUI struct {
	connUI *connUI
	dbName string
	db     *sql.DB

	tableList *duit.List
	viewUI    *duit.Box // holds 1 kid, the editUI, tableUI or placeholder label

	*duit.Box // holds either box with status message, or box with tableList and viewUI
}

func newDBUI(cUI *connUI, dbName string) (ui *dbUI) {
	ui = &dbUI{
		connUI: cUI,
		dbName: dbName,
	}
	ui.Box = &duit.Box{}
	return
}

func (ui *dbUI) layout() {
	dui.MarkLayout(nil) // xxx
}

func (ui *dbUI) error(err error) {
	defer ui.layout()
	msg := &duit.Label{Text: "error: " + err.Error()}
	retry := &duit.Button{
		Text: "retry",
		Click: func(e *duit.Event) {
			go ui.init()
		},
	}
	ui.Box.Kids = duit.NewKids(middle(msg, retry))
}

func (ui *dbUI) init() {
	setError := func(err error) {
		dui.Call <- func() {
			ui.error(err)
		}
	}

	db, err := sql.Open("postgres", ui.connUI.cc.connectionString(ui.dbName))
	if err != nil {
		setError(err)
		return
	}

	ctx, cancelQueryFunc := context.WithTimeout(context.Background(), 15*time.Second)
	dui.Call <- func() {
		defer ui.layout()
		ui.Box.Kids = duit.NewKids(
			middle(
				&duit.Label{Text: "listing tables..."},
				&duit.Button{
					Text: "cancel",
					Click: func(e *duit.Event) {
						cancelQueryFunc()
					},
				},
			),
		)
	}
	q := `
		select coalesce(json_agg(x.name order by internal asc, name asc), '[]') from (
			select table_schema in ('pg_catalog', 'information_schema') as internal, table_schema || '.' || table_name as name
			from information_schema.tables
		) x
	`
	var tables []string
	err = parseRow(db.QueryRowContext(ctx, q), &tables, "listing tables in database")
	if err != nil {
		setError(err)
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
