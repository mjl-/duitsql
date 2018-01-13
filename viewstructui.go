package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"image"

	"github.com/mjl-/duit"
)

type viewstructUI struct {
	dbUI      *dbUI
	name      string
	vertical  *duit.Vertical
	scrollBox *duit.Box
	duit.Box
}

func newViewStructUI(dbUI *dbUI, name string) *viewstructUI {
	scrollBox := &duit.Box{}
	scroll := &duit.Scroll{
		Kid: duit.Kid{
			UI: scrollBox,
		},
	}
	vertical := &duit.Vertical{
		Split: func(height int) []int {
			return []int{height / 2, height - height/2}
		},
		Kids: duit.NewKids(scroll, nil), // nil will be replaced by an Edit
	}
	return &viewstructUI{
		dbUI:      dbUI,
		name:      name,
		vertical:  vertical,
		scrollBox: scrollBox,
	}
}

func (ui *viewstructUI) layout() {
	dui.MarkLayout(ui)
}

func (ui *viewstructUI) status(msg string) {
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
func (ui *viewstructUI) init() {
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
func (ui *viewstructUI) _load(ctx context.Context, cancelQueryFunc func()) {
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

	var qDefinition, qColumns string
	var args []interface{}
	switch ui.dbUI.connUI.cc.Type {
	case "postgres":
		qDefinition = `
			select view_definition
			from information_schema.views
			where table_schema || '.' || table_name = $1
		`
		qColumns = `
			select
				column_name as name,
				udt_name as type
			from information_schema.columns
			where table_schema || '.' || table_name=$1
			order by ordinal_position
		`
		args = append(args, ui.name)
	case "mysql":
		qDefinition = `
			select view_definition
			from information_schema.views
			where table_schema=? and table_name=?
		`
		// todo: more complete type, eg size to varchar and ints and such
		qColumns = `
			select
				column_name as name,
				data_type as type
			from information_schema.columns
			where table_schema=? and table_name=?
			order by ordinal_position
		`
		args = append(args, ui.dbUI.dbName, ui.name)
	case "sqlserver":
		qDefinition = `
			select view_definition
			from information_schema.views
			where concat(table_schema, '.', table_name)=@name
		`
		// todo: more complete type, eg size to varchar and ints and such
		qColumns = `
			select
				column_name as name,
				data_type as type
			from information_schema.columns
			where concat(table_schema, '.', table_name)=@name
			order by ordinal_position
		`
		args = append(args, sql.Named("name", ui.name))
	default:
		panic("bad connection type")
	}

	var definition sql.NullString
	err := ui.dbUI.db.QueryRowContext(ctx, qDefinition, args...).Scan(&definition)
	lcheck(err, "fetching view definition")

	type column struct {
		Name string
		Type string
	}
	var columns []column
	rows, err := ui.dbUI.db.QueryContext(ctx, qColumns, args...)
	lcheck(err, "fetching columns")
	defer rows.Close()
	for rows.Next() {
		var col column
		err = rows.Scan(&col.Name, &col.Type)
		lcheck(err, "scanning row")
		columns = append(columns, col)
	}
	lcheck(rows.Err(), "reading row")

	var columnUIs []duit.UI
	for _, e := range columns {
		columnUIs = append(columnUIs,
			label(e.Name),
			label(e.Type),
		)
	}

	edit := duit.NewEdit(bytes.NewReader([]byte(definition.String)))

	dui.Call <- func() {
		ui.scrollBox.Padding = duit.Space{Top: duit.ScrollbarSize, Right: duit.ScrollbarSize, Bottom: 6, Left: duit.ScrollbarSize}
		ui.scrollBox.Margin = image.Pt(0, 6)
		ui.scrollBox.Kids = duit.NewKids(
			&duit.Label{Font: bold, Text: "columns"},
			&duit.Grid{
				Columns: 2,
				Width:   -1,
				Padding: []duit.Space{
					duit.Space{Top: 1, Right: 4, Bottom: 1, Left: 0},
					duit.Space{Top: 1, Right: 0, Bottom: 2, Left: 4},
				},
				Kids: duit.NewKids(columnUIs...),
			},
			&duit.Label{Font: bold, Text: "definition"},
		)
		ui.vertical.Kids[1].UI = edit
		ui.Box.Kids = duit.NewKids(ui.vertical)
		ui.layout()
	}
}
