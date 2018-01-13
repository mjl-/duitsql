package main

import (
	"context"
	"database/sql"
	"fmt"
	"image"

	"github.com/mjl-/duit"
)

type tablestructUI struct {
	dbUI *dbUI
	name string
	duit.Box
}

func newTableStructUI(dbUI *dbUI, name string) *tablestructUI {
	return &tablestructUI{
		dbUI: dbUI,
		name: name,
	}
}

func (ui *tablestructUI) layout() {
	dui.MarkLayout(ui)
}

func (ui *tablestructUI) status(msg string) {
	retry := &duit.Button{
		Text: "retry",
		Click: func(e *duit.Event) {
			ui.init()
		},
	}
	ui.Box.Kids = duit.NewKids(middle(label(msg), retry))
	ui.layout()
}

// called from main loop
func (ui *tablestructUI) init() {
	ctx, cancelQueryFunc := context.WithCancel(context.Background())

	cancel := &duit.Button{
		Text: "cancel",
		Click: func(e *duit.Event) {
			cancelQueryFunc()
		},
	}
	ui.Box.Kids = duit.NewKids(middle(label("executing query..."), cancel))
	ui.layout()

	go ui._load(ctx, cancelQueryFunc)
}

// called from outside main loop
func (ui *tablestructUI) _load(ctx context.Context, cancelQueryFunc func()) {
	defer cancelQueryFunc()

	var lerr bool
	defer func() {
		e := recover()
		if lerr {
			return
		}
		if e != nil {
			panic(e)
		}
	}()
	lcheck := func(err error, msg string) {
		if err == nil {
			return
		}
		dui.Call <- func() {
			ui.status(fmt.Sprintf("error: %s: %s", msg, err))
		}
		lerr = true
		panic(lerr)
	}

	var qColumns string
	var args []interface{}
	switch ui.dbUI.connUI.cc.Type {
	case "postgres":
		qColumns = `
			select
				column_name as name,
				udt_name as type,
				column_default as ddefault_valueefault,
				is_nullable = 'YES' as isnullable
			from information_schema.columns
			where table_schema || '.' || table_name=$1
			order by ordinal_position
		`
		args = append(args, ui.name)
	case "mysql":
		qColumns = `
			select
				column_name as name,
				data_type as type,
				column_default as default_value,
				is_nullable = 'YES' as isnullable
			from information_schema.columns
			where table_schema=? and table_name=?
			order by ordinal_position
		`
		args = append(args, ui.dbUI.dbName, ui.name)
	case "sqlserver":
		qColumns = `
			select
				column_name as name,
				data_type as type,
				column_default as default_value,
				case when is_nullable = 'YES' then 1 else 0 end as isnullable
			from information_schema.columns
			-- where concat(table_schema, '.', table_name)=@name
			order by ordinal_position
		`
		args = append(args, sql.Named("name", ui.name))
	default:
		panic("bad connection type")
	}
	type column struct {
		Name         sql.NullString
		Type         sql.NullString
		DefaultValue sql.NullString
		IsNullable   bool
	}
	var columns []column
	rows, err := ui.dbUI.db.QueryContext(ctx, qColumns, args...)
	lcheck(err, "fetching columns")
	defer rows.Close()
	for rows.Next() {
		var col column
		err = rows.Scan(&col.Name, &col.Type, &col.DefaultValue, &col.IsNullable)
		lcheck(err, "scanning row")
		columns = append(columns, col)
	}
	lcheck(rows.Err(), "reading row")

	columnUIs := []duit.UI{
		&duit.Label{Font: bold, Text: "name"},
		&duit.Label{Font: bold, Text: "type"},
		&duit.Label{Font: bold, Text: "default"},
		&duit.Label{Font: bold, Text: "nullable"},
	}
	for _, e := range columns {
		nullable := "NOT NULL"
		if e.IsNullable {
			nullable = "NULL"
		}
		columnUIs = append(columnUIs,
			label(e.Name.String),
			label(e.Type.String),
			label(e.DefaultValue.String),
			label(nullable),
		)
	}

	dui.Call <- func() {
		ui.Box.Padding = duit.Space{Top: duit.ScrollbarSize, Right: duit.ScrollbarSize, Bottom: 6, Left: duit.ScrollbarSize}
		ui.Box.Margin = image.Pt(0, 6)
		ui.Box.Kids = duit.NewKids(
			&duit.Label{Font: bold, Text: "columns"},
			&duit.Grid{
				Columns: 4,
				Width:   -1,
				Padding: []duit.Space{
					duit.Space{Top: 1, Right: 4, Bottom: 1, Left: 0},
					duit.Space{Top: 1, Right: 2, Bottom: 1, Left: 2},
					duit.Space{Top: 1, Right: 2, Bottom: 1, Left: 2},
					duit.Space{Top: 1, Right: 0, Bottom: 1, Left: 4},
				},
				Kids: duit.NewKids(columnUIs...),
			},
		)
		ui.layout()
	}
}
