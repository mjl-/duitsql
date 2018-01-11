package main

import (
	"fmt"
	"reflect"
	"context"

	"github.com/mjl-/duit"
)

type tableUI struct {
	dbUI  *dbUI
	query string

	grid *duit.Gridlist

	*duit.Box
}

func newTableUI(dbUI *dbUI, query string) *tableUI {
	ui := &tableUI{
		dbUI:  dbUI,
		query: query,
		Box: &duit.Box{},
	}
	return ui
}

func (ui *tableUI) layout() {
	dui.MarkLayout(nil) // xxx
}

func (ui *tableUI) error(err error) {
	defer ui.layout()
	msg := &duit.Label{Text: fmt.Sprintf("error: %s", err)}
	retry := &duit.Button{
		Text: "retry",
		Click: func(e *duit.Event) {
			go ui.load()
		},
	}
	ui.Box.Kids = duit.NewKids(middle(msg, retry))
}

func (ui *tableUI) load() {
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
			ui.error(fmt.Errorf("%s: %s", msg, err))
		}
		lerr = true
		panic(lerr)
	}

	ctx, cancelQueryFunc := context.WithCancel(context.Background())
	dui.Call <- func() {
		msg := &duit.Label{Text: "executing query..."}
		cancel := &duit.Button{
			Text: "cancel",
			Click: func(e *duit.Event) {
				cancelQueryFunc()
			},
		}
		ui.Box.Kids = duit.NewKids(middle(msg, cancel))
		ui.layout()
	}

	rows, err := ui.dbUI.db.QueryContext(ctx, ui.query)
	lcheck(err, "fetching table")

	colNames, err := rows.Columns()
	lcheck(err, "reading column names")

	colTypes, err := rows.ColumnTypes()
	lcheck(err, "reading column types")

	halign := make([]duit.Halign, len(colTypes))

	vals := make([]interface{}, len(colTypes))
	for i, t := range colTypes {
		tt := t.ScanType()
		vals[i] = reflect.New(reflect.PtrTo(tt)).Interface()
		if tt.Kind() == reflect.String || len(colTypes) == 1 {
			halign[i] = duit.HalignLeft
		} else {
			halign[i] = duit.HalignRight
		}
	}
	gridRows := []*duit.Gridrow{}
	for rows.Next() {
		err = rows.Scan(vals...)
		lcheck(err, "scanning row")
		l := make([]string, len(vals))
		for i, v := range vals {
			vv := reflect.ValueOf(v)
			if vv.IsNil() || vv.Elem().IsNil() {
				l[i] = "NULL"
			} else {
				l[i] = fmt.Sprintf("%v", vv.Elem().Elem().Interface())
			}
		}
		gridRow := &duit.Gridrow{
			Values: l,
		}
		gridRows = append(gridRows, gridRow)
	}
	err = rows.Err()
	lcheck(err, "reading next row")

	dui.Call <- func() {
		if len(gridRows) == 0 {
			ui.error(fmt.Errorf("empty resultset"))
			return
		}

		ui.grid = &duit.Gridlist{
			Header:   duit.Gridrow{Values: colNames},
			Rows:     gridRows,
			Halign:   halign,
			Multiple: true,
			Striped:  true,
			Padding:  duit.SpaceXY(4, 4),
		}
		ui.Box.Kids = duit.NewKids(duit.NewScroll(ui.grid))
		ui.layout()
	}
}
