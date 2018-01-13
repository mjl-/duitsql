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
	dbUI *dbUI
	name string
	duit.Box
}

func newViewStructUI(dbUI *dbUI, name string) *viewstructUI {
	return &viewstructUI{
		dbUI: dbUI,
		name: name,
	}
}

func (ui *viewstructUI) layout() {
	dui.MarkLayout(ui)
}

func (ui *viewstructUI) status(msg string) {
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
func (ui *viewstructUI) init() {
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

	qView := `
		select view_definition
		from information_schema.views
		where table_schema || '.' || table_name = $1
	`
	var definition sql.NullString
	err := ui.dbUI.db.QueryRowContext(ctx, qView, ui.name).Scan(&definition)
	lcheck(err, "fetching view definition")

	qColumns := `
		select coalesce(json_agg(x.*), '[]')
		from (
			select
				column_name as name,
				udt_name as type
			from information_schema.columns
			where table_schema || '.' || table_name=$1
			order by ordinal_position
		) x
	`
	type column struct {
		Name string
		Type string
	}
	var columns []column
	err = parseRow(ui.dbUI.db.QueryRowContext(ctx, qColumns, ui.name), &columns)
	lcheck(err, "fetching columns")

	var columnUIs []duit.UI
	for _, e := range columns {
		columnUIs = append(columnUIs,
			&duit.Label{Text: e.Name},
			&duit.Label{Text: e.Type},
		)
	}

	edit := duit.NewEdit(bytes.NewReader([]byte(definition.String)))

	dui.Call <- func() {
		ui.Box.Kids = duit.NewKids(
			&duit.Box{
				Padding: duit.Space{Top: duit.ScrollbarSize, Right: duit.ScrollbarSize, Bottom: 6, Left: duit.ScrollbarSize},
				Margin:  image.Pt(0, 6),
				Kids: duit.NewKids(
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
				),
			},
			edit,
		)
		ui.layout()
	}
}
