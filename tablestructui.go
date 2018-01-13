package main

import (
	"context"
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
	label := &duit.Label{Text: msg}
	retry := &duit.Button{
		Text: "retry",
		Click: func(e *duit.Event) {
			ui.init()
		},
	}
	ui.Box.Kids = duit.NewKids(middle(label, retry))
	ui.layout()
}

// called from main loop
func (ui *tablestructUI) init() {
	ctx, cancelQueryFunc := context.WithCancel(context.Background())

	msg := &duit.Label{Text: "executing query..."}
	cancel := &duit.Button{
		Text: "cancel",
		Click: func(e *duit.Event) {
			cancelQueryFunc()
		},
	}
	ui.Box.Kids = duit.NewKids(middle(msg, cancel))
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

	qColumns := `
		select coalesce(json_agg(x.*), '[]')
		from (
			select
				column_name as name,
				udt_name as type,
				column_default as default,
				is_nullable = 'YES' as isnullable
			from information_schema.columns
			where table_schema || '.' || table_name=$1
			order by ordinal_position
		) x
	`
	type column struct {
		Name       string
		Type       string
		Default    string
		IsNullable bool
	}
	var columns []column
	err := parseRow(ui.dbUI.db.QueryRowContext(ctx, qColumns, ui.name), &columns)
	lcheck(err, "fetching columns")

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
			&duit.Label{Text: e.Name},
			&duit.Label{Text: e.Type},
			&duit.Label{Text: e.Default},
			&duit.Label{Text: nullable},
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
