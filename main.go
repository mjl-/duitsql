package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"path"

	"9fans.net/go/draw"
	_ "github.com/lib/pq"
	"github.com/mjl-/duit"
)

type configConnection struct {
	Name     string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	TLS      bool
}

func (cc configConnection) connectionString(dbName string) string {
	tls := "disable"
	if cc.TLS {
		tls = "verify-full"
	}
	s := fmt.Sprintf("user=%s password=%s host=%s port=%d sslmode=%s application_name=duitsql", cc.User, cc.Password, cc.Host, cc.Port, tls)
	if dbName != "" {
		s += fmt.Sprintf(" dbname=%s", dbName)
	}
	return s
}

var (
	dui           *duit.DUI
	connections   *duit.List
	connectionBox *duit.Box
	disconnect    *duit.Button
	hideLeftBars  bool // whether to show connections & databases bar. if not, we make them zero width
	bold          *draw.Font
)

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s\n", msg, err)
	}
}

func saveConfigConnections(l []configConnection) {
	p := os.Getenv("HOME") + "/lib/duit/sql/connections.json"
	os.MkdirAll(path.Dir(p), 0777)
	f, err := os.Create(p)
	if err == nil {
		err = json.NewEncoder(f).Encode(l)
	}
	if f != nil {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}
	if err != nil {
		log.Printf("saving config: %s\n", err)
	}
}

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		log.Println("usage: duitsql")
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) != 0 {
		flag.Usage()
		os.Exit(2)
	}

	var err error
	dui, err = duit.NewDUI("sql", "1000x600")
	check(err, "new dui")

	bold = dui.Display.DefaultFont
	if os.Getenv("boldfont") != "" {
		bold, err = dui.Display.OpenFont(os.Getenv("boldfont"))
		check(err, "open bold font")
	}

	var configConnections []configConnection
	f, err := os.Open(os.Getenv("HOME") + "/lib/duit/sql/connections.json")
	if err != nil && !os.IsNotExist(err) {
		check(err, "opening connections.json config file")
	}
	if f != nil {
		err = json.NewDecoder(f).Decode(&configConnections)
		check(err, "parsing connections.json config file")
		check(f.Close(), "closing connections.json config file")
	}

	noConnectionUI := duit.NewMiddle(&duit.Label{Text: "select a connection on the left"})
	connectionBox = &duit.Box{
		Kids: duit.NewKids(noConnectionUI),
	}

	connectionValues := make([]*duit.ListValue, len(configConnections)+1)
	for i, cc := range configConnections {
		lv := &duit.ListValue{Text: cc.Name, Value: newConnUI(cc)}
		connectionValues[i] = lv
	}
	connectionValues[len(connectionValues)-1] = &duit.ListValue{Text: "<new>", Value: nil}

	connections = &duit.List{
		Values: connectionValues,
		Changed: func(index int, r *duit.Event) {
			defer dui.MarkLayout(connectionBox)
			disconnect.Disabled = true
			lv := connections.Values[index]
			if !lv.Selected {
				connectionBox.Kids = duit.NewKids(noConnectionUI)
				return
			}
			if lv.Value == nil {
				connectionBox.Kids = duit.NewKids(newSettingsUI(configConnection{}, func() {}))
				return
			}
			cUI := lv.Value.(*connUI)
			disconnect.Disabled = cUI.db == nil
			connectionBox.Kids = duit.NewKids(cUI)
		},
	}

	toggleSlim := &duit.Button{
		Text: "toggle left",
		Click: func(r *duit.Event) {
			hideLeftBars = !hideLeftBars
			dui.MarkLayout(nil)
		},
	}
	disconnect = &duit.Button{
		Text:     "disconnect",
		Disabled: true,
		Click: func(r *duit.Event) {
			dui.MarkLayout(nil)
			l := connections.Selected()
			if len(l) != 1 {
				return
			}
			lv := connections.Values[l[0]]
			cUI := lv.Value.(*connUI)
			cUI.disconnect()
		},
	}
	status := &duit.Label{}

	dui.Top.UI = &duit.Box{
		Kids: duit.NewKids(
			&duit.Box{
				Padding: duit.SpaceXY(4, 2),
				Margin:  image.Pt(4, 0),
				Kids:    duit.NewKids(toggleSlim, disconnect, status),
			},
			&duit.Horizontal{
				Split: func(width int) []int {
					if hideLeftBars {
						return []int{0, width}
					}
					first := dui.Scale(125)
					if first > width/2 {
						first = width / 2
					}
					return []int{first, width - first}
				},
				Kids: duit.NewKids(
					&duit.Box{
						Kids: duit.NewKids(
							duit.CenterUI(duit.SpaceXY(4, 2), &duit.Label{Text: "connections", Font: bold}),
							duit.NewScroll(connections),
						),
					},
					connectionBox,
				),
			},
		),
	}
	dui.Render()

	for {
		select {
		case e := <-dui.Inputs:
			dui.Input(e)

		case <-dui.Done:
			return
		}
	}
}

func parseRow(row *sql.Row, r interface{}, msg string) error {
	var buf []byte
	err := row.Scan(&buf)
	if err != nil {
		return fmt.Errorf("%s: scanning json bytes from database: %s", msg, err)
	}
	err = json.Unmarshal(buf, r)
	if err != nil {
		return fmt.Errorf("%s: parsing json from database: %s", msg, err)
	}
	return nil
}
