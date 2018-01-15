package main

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/mjl-/duit"
)

type resultUI struct {
	dbUI  *dbUI
	query string
	grid  *duit.Gridlist

	duit.Box
}

func newResultUI(dbUI *dbUI, query string) *resultUI {
	ui := &resultUI{
		dbUI:  dbUI,
		query: query,
	}
	return ui
}

func (ui *resultUI) layout() {
	dui.MarkLayout(nil) // xxx
}

func (ui *resultUI) status(msg string) {
	defer ui.layout()
	retry := &duit.Button{
		Text: "retry",
		Click: func() (e duit.Event) {
			go ui.load()
			return
		},
	}
	ui.Box.Kids = duit.NewKids(middle(label(msg), retry))
}

func (ui *resultUI) load() {
	lcheck, handle := errorHandler(func(err error) {
		dui.Call <- func() {
			ui.status(fmt.Sprintf("error: %s", err))
		}
	})
	defer handle()

	status := label("executing query...")
	ctx, cancelQueryFunc := context.WithCancel(context.Background())
	defer cancelQueryFunc()
	dui.Call <- func() {
		cancel := &duit.Button{
			Text: "cancel",
			Click: func() (e duit.Event) {
				cancelQueryFunc()
				return
			},
		}
		ui.Box.Kids = duit.NewKids(middle(status, cancel))
		ui.layout()
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		n := 0
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				n++
				dui.Call <- func() {
					status.Text = fmt.Sprintf("executing query... %ds", n)
					ui.layout()
				}
			}
		}
	}()

	rows, err := ui.dbUI.db.QueryContext(ctx, ui.query)
	lcheck(err, "executing query")
	defer rows.Close()

	colNames, err := rows.Columns()
	lcheck(err, "reading column names")

	colTypes, err := rows.ColumnTypes()
	lcheck(err, "reading column types")

	halign := make([]duit.Halign, len(colTypes))

	vals := make([]interface{}, len(colTypes))
	isBinary := make([]bool, len(colTypes))
	for i, t := range colTypes {
		isBinary[i] = t.DatabaseTypeName() == "BYTEA" // postgres
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
				v := vv.Elem().Elem().Interface()
				// log.Printf("value %#v, %T\n", v, v)
				if vv, ok := v.([]byte); ok {
					v = string(vv)
					if isBinary[i] {
						v = fmt.Sprintf("%x", v)
					}
				}
				l[i] = fmt.Sprintf("%v", v)
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
			ui.status(fmt.Sprintf("empty resultset"))
			return
		}

		ui.grid = &duit.Gridlist{
			Header:   &duit.Gridrow{Values: colNames},
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
