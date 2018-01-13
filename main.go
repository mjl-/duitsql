package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"log"
	"net/url"
	"os"
	"path"
	"strings"

	"9fans.net/go/draw"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/mjl-/duit"
)

type configConnection struct {
	Type     string // postgres, mysql, sqlserver
	Name     string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	TLS      bool
}

func (cc configConnection) connectionString(dbName string) string {
	switch cc.Type {
	case "postgres":
		tls := "disable"
		if cc.TLS {
			tls = "verify-full"
		}
		s := fmt.Sprintf("user=%s password=%s host=%s port=%d sslmode=%s application_name=duitsql", cc.User, cc.Password, cc.Host, cc.Port, tls)
		if dbName != "" {
			s += fmt.Sprintf(" dbname=%s", dbName)
		}
		return s
	case "mysql":
		s := ""
		if cc.User != "" || cc.Password != "" {
			s += cc.User
			if cc.Password != "" {
				s += ":" + cc.Password
			}
			s += "@"
		}
		address := cc.Host
		if cc.Port != 0 {
			address += fmt.Sprintf(":%d", cc.Port)
		}
		s += fmt.Sprintf("tcp(%s)", address)
		s += "/"
		if dbName != "" {
			s += dbName
		}
		if cc.TLS {
			s += "?tls=true"
		}
		return s
	case "sqlserver":
		host := cc.Host
		if cc.Port != 0 {
			host += fmt.Sprintf(":%d", cc.Port)
		}
		qs := []string{}
		if dbName != "" {
			qs = append(qs, "database="+url.QueryEscape(dbName))
		}
		if cc.TLS {
			qs = append(qs, "encrypt=true", "TrustServerCertificate=false")
		}
		u := &url.URL{
			Scheme:   "sqlserver",
			User:     url.UserPassword(cc.User, cc.Password),
			Host:     host,
			RawQuery: strings.Join(qs, "&"),
		}
		return u.String()
	}
	panic("missing case")
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
	p := os.Getenv("HOME") + "/lib/duitsql/connections.json"
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
	f, err := os.Open(os.Getenv("HOME") + "/lib/duitsql/connections.json")
	if err != nil && !os.IsNotExist(err) {
		check(err, "opening connections.json config file")
	}
	if f != nil {
		err = json.NewDecoder(f).Decode(&configConnections)
		check(err, "parsing connections.json config file")
		check(f.Close(), "closing connections.json config file")
	}
	for _, cc := range configConnections {
		switch cc.Type {
		case "postgres", "mysql", "sqlserver":
		default:
			log.Fatalf("unknown connection type %q\n", cc.Type)
		}
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
				connectionBox.Kids = duit.NewKids(newSettingsUI(configConnection{Type: "postgres"}, func() {}))
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

func parseRow(row *sql.Row, r interface{}) error {
	var buf []byte
	err := row.Scan(&buf)
	if err != nil {
		return fmt.Errorf("scanning json bytes from database: %s", err)
	}
	err = json.Unmarshal(buf, r)
	if err != nil {
		return fmt.Errorf("parsing json from database: %s", err)
	}
	return nil
}
