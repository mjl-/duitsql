package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
)

type editUI struct {
	dbUI *dbUI

	sqlPath   string
	edit      *duit.Edit
	resultBox *duit.Box

	duit.Box
}

func (ui *editUI) layout() {
	dui.MarkLayout(nil) // xxx
}

func newEditUI(dbUI *dbUI) (ui *editUI) {
	sqlPath := fmt.Sprintf("%s/lib/duit/sql/%s.%s.sql", os.Getenv("HOME"), dbUI.connUI.cc.Name, dbUI.dbName)

	edit := &duit.Edit{}
	// xxx todo: do not read in main loop?
	sqlF, err := os.Open(sqlPath)
	if err == nil {
		buf, err := ioutil.ReadAll(sqlF)
		if err == nil {
			edit = duit.NewEdit(bytes.NewReader(buf))
		}
		sqlF.Close()
	}
	edit.Keys = func(k rune, m draw.Mouse) (e duit.Event) {
		switch k {
		case draw.KeyCmd + 'g':
			log.Printf("executing command\n")
			query := ui.edit.Selection()
			if query == "" {
				// read query under cursor. by moving backward and forward until we find eof or ;
				c := ui.edit.Cursor()
				skip := func(r duit.EditReader) {
					for {
						c, eof := r.Peek()
						if eof || c == ';' {
							break
						}
						r.Get()
					}
				}
				br := ui.edit.ReverseEditReader(c)
				fr := ui.edit.EditReader(c)
				skip(br)
				skip(fr)

				buf, err := ioutil.ReadAll(io.NewSectionReader(ui.edit.Reader(), br.Offset(), fr.Offset()-br.Offset()))
				if err != nil {
					log.Printf("error reading from edit: %s\n", err)
					return
				}
				query = string(buf)
			}
			log.Printf("query is %q\n", query)
			defer ui.layout()
			tabUI := newResultUI(ui.dbUI, query)
			ui.resultBox.Kids = duit.NewKids(tabUI)
			go tabUI.load()

		case draw.KeyCmd + 's':
			os.MkdirAll(path.Dir(ui.sqlPath), 0777)
			f, err := os.Create(ui.sqlPath)
			if err == nil {
				_, err = io.Copy(f, ui.edit.Reader())
			}
			if f != nil {
				err2 := f.Close()
				if err == nil {
					err = err2
				}
			}
			if err != nil {
				log.Printf("write sql: %s\n", err)
			}
		default:
			return
		}
		e.Consumed = true
		return
	}
	resultBox := &duit.Box{
		Kids: duit.NewKids(
			duit.NewMiddle(
				duit.SpaceXY(10, 10),
				label("type a query and execute selection or query under cursor with cmd + g"),
			),
		),
	}
	ui = &editUI{
		dbUI:      dbUI,
		sqlPath:   sqlPath,
		edit:      edit,
		resultBox: resultBox,
		Box: duit.Box{
			Kids: duit.NewKids(
				&duit.Split{
					Vertical: true,
					Gutter:   1,
					Split: func(height int) []int {
						half := height / 2
						return []int{half, height - half}
					},
					Kids: duit.NewKids(edit, resultBox),
				},
			),
		},
	}
	ui.Box.Kids[0].ID = "edit"
	return
}
