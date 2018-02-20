/*
UI structure:
One mainUI as variable topUI. It has a list of connections and possibly an active connUI. Instead of the active connUI, we can always have a settingsUI in its place.
A connUI has a list of databases and possibly an active dbUI.
A dbUI has a list of tables/views and possibly an active tableUI, viewUI or editUI.
A tableUI and viewUI are very similar: they have a Tabs to switch between rows (resultUI) and structure view (tablestructUI/viewstructUI).
An editUI has a duit.Edit for the SQL, and a resultUI.
*/

/*
SQL database browser and query executor, created with duit.

Duitsql can connect to PostgreSQL, MySQL and SQLServer servers through their pure Go database drivers. Only these provide basic SQL introspection through an information_schema.

Using

Select and manage connections to database servers on the left (type/user/password/host/port).
Select a database, then a table/view or write your own SQL query.
You will see the rows in the selected table/view or the query results. You can also choose to view the structure of the database objects (columns and types, etc).

Connections are stored in $appdata/duitsql/connections.json, including passwords.
SQL scripts are stored in $appdata/duitsql/$connectionname.$databasename.sql.
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"9fans.net/go/draw"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/mjl-/duit"
)

var (
	dui   *duit.DUI
	bold  *draw.Font
	topUI *mainUI
)

func check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s\n", msg, err)
	}
}

func connectionsJSONPath() string {
	return duit.AppDataDir("duitsql") + "/connections.json"
}

func saveConnectionConfigs(l []connectionConfig) {
	lcheck, handle := errorHandler(func(err error) {
		dui.Call <- func() {
			topUI.status.Text = fmt.Sprintf("saving config: %s\n", err)
			dui.MarkLayout(topUI.status)
		}
	})
	defer handle()
	p := connectionsJSONPath()
	os.MkdirAll(path.Dir(p), 0777)
	f, err := os.Create(p)
	lcheck(err, "create")
	err = json.NewEncoder(f).Encode(l)
	lcheck(err, "encoding json")
	err = f.Close()
	lcheck(err, "close")
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
	dui, err = duit.NewDUI("sql", nil)
	check(err, "new dui")

	bold = dui.Display.DefaultFont
	if os.Getenv("fontbold") != "" {
		bold, err = dui.Display.OpenFont(os.Getenv("fontbold"))
		check(err, "open bold font")
	}

	var configs []connectionConfig
	f, err := os.Open(connectionsJSONPath())
	if err != nil && !os.IsNotExist(err) {
		check(err, "opening connections.json config file")
	}
	if f != nil {
		err = json.NewDecoder(f).Decode(&configs)
		check(err, "parsing connections.json config file")
		check(f.Close(), "closing connections.json config file")
	}
	for _, c := range configs {
		switch c.Type {
		case "postgres", "mysql", "sqlserver":
		default:
			log.Fatalf("unknown connection type %q\n", c.Type)
		}
	}

	topUI = newMainUI(configs)
	dui.Top.UI = topUI
	dui.Render()

	for {
		select {
		case e := <-dui.Inputs:
			dui.Input(e)

		case err, ok := <-dui.Error:
			if !ok {
				return
			}
			log.Printf("duit: %s\n", err)
		}
	}
}
