package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mjl-/duit"
)

type dbUI struct {
	connUI *connUI
	dbName string
	db     *sql.DB

	tableList *duit.Gridlist
	contentUI *duit.Box // holds 1 kid, the editUI, tableUI, viewUI or placeholder label

	duit.Box // holds either box with status message, or box with tableList and contentUI
}

func newDBUI(cUI *connUI, dbName string) (ui *dbUI) {
	ui = &dbUI{
		connUI: cUI,
		dbName: dbName,
	}
	return
}

func (ui *dbUI) layout() {
	dui.MarkLayout(nil) // xxx
}

func (ui *dbUI) status(msg string) {
	retry := &duit.Button{
		Text: "retry",
		Click: func() (e duit.Event) {
			go ui.init()
			return
		},
	}
	ui.Box.Kids = duit.NewKids(middle(label(msg), retry))
	ui.layout()
}

// called from outside of main context
func (ui *dbUI) init() {
	lcheck, handle := errorHandler(func(err error) {
		dui.Call <- func() {
			ui.status(fmt.Sprintf("error: %s", err))
		}
	})
	defer handle()

	db, err := sql.Open(ui.connUI.config.Type, ui.connUI.config.connectionString(ui.dbName))
	lcheck(err, "connecting to database")

	ctx, cancelQueryFunc := context.WithTimeout(context.Background(), 15*time.Second)
	dui.Call <- func() {
		cancel := &duit.Button{
			Text: "cancel",
			Click: func() (e duit.Event) {
				cancelQueryFunc()
				return
			},
		}
		ui.Box.Kids = duit.NewKids(middle(label("listing tables..."), cancel))
		ui.layout()
	}

	var q string
	var args []interface{}
	switch ui.connUI.config.Type {
	case "postgres":
		q = `
			select
				table_type = 'VIEW' as is_view,
				table_schema || '.' || table_name as name
			from information_schema.tables
			order by table_schema in ('pg_catalog', 'information_schema') asc, name asc
		`
	case "mysql":
		// in mysql, schema & database are the same concept, so no need to add the schema to the name here
		q = `
			select
				table_type like '%VIEW' as is_view,
				table_name as name
			from information_schema.tables
			where table_schema = ?
			order by name asc
		`
		args = append(args, ui.dbName)
	case "sqlserver":
		q = `
			select
				case table_type when 'VIEW' then 1 else 0 end,
				concat(table_schema, '.', table_name) as name
			from information_schema.tables
			order by name
		`
	default:
		panic("bad connection type")
	}
	type object struct {
		IsView bool   `json:"is_view"`
		Name   string `json:"name"`
	}
	var objects []object
	rows, err := db.QueryContext(ctx, q, args...)
	lcheck(err, "listing tables and views")
	defer rows.Close()
	for rows.Next() {
		var o object
		err = rows.Scan(&o.IsView, &o.Name)
		lcheck(err, "scanning row")
		objects = append(objects, o)
	}
	lcheck(rows.Err(), "reading row")

	eUI := newEditUI(ui)
	values := make([]*duit.Gridrow, 1+len(objects))
	values[0] = &duit.Gridrow{
		Selected: true,
		Values:   []string{"", "<sql>"},
		Value:    eUI,
	}
	for i, obj := range objects {
		var objUI duit.UI
		var kind string
		if obj.IsView {
			objUI = newViewUI(ui, obj.Name)
			kind = "V "
		} else {
			objUI = newTableUI(ui, obj.Name)
			kind = "T "
		}
		values[i+1] = &duit.Gridrow{
			Values: []string{
				kind,
				obj.Name,
			},
			Value: objUI,
		}
	}

	dui.Call <- func() {
		defer ui.layout()
		ui.db = db
		ui.tableList = &duit.Gridlist{
			Fit:    duit.FitSlim,
			Halign: []duit.Halign{duit.HalignMiddle, duit.HalignLeft},
			Rows:   values,
			Changed: func(index int) (e duit.Event) {
				lv := ui.tableList.Rows[index]
				var selUI, focusUI duit.UI
				if !lv.Selected {
					selUI = middle(label("select <sql>, or a a table or view on the left"))
				} else {
					selUI = lv.Value.(duit.UI)
					switch objUI := selUI.(type) {
					case *editUI:
						focusUI = objUI.edit
					case *tableUI:
						objUI.init()
						focusUI = objUI.tabsUI.Buttongroup
					case *viewUI:
						objUI.init()
						focusUI = objUI.tabsUI.Buttongroup
					}
				}
				ui.contentUI.Kids = duit.NewKids(selUI)
				ui.layout()
				if focusUI != nil {
					dui.Render()
					dui.Focus(focusUI)
				}
				return
			},
		}
		ui.contentUI = &duit.Box{
			Kids: duit.NewKids(eUI),
		}
		ui.Box.Kids = duit.NewKids(
			&duit.Split{
				Gutter:     1,
				Background: dui.Gutter,
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
					ui.contentUI,
				),
			},
		)
		ui.Box.Kids[0].ID = "tables"
	}
}
