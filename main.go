/*
SQL database browser and query executor, created with duit.

Duitsql can connect to PostgreSQL, MySQL and MS SQLServer. All through their pure Go database drivers, and because they have an information_schema describing their table structure.
Select and manage connections to database server on the left (type/user/password/host/port).
Next select one of the database, then one of the tables or manual SQL input.
Finally view the tables/views, the query results, or the structure of the objects.
*/
package main

import (
	"encoding/json"
	"flag"
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

func saveConfigConnections(l []configConnection) {
	lcheck, handle := errorHandler(func(err error) {
		log.Printf("saving config: %s\n", err)
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

	var configConnections []configConnection
	f, err := os.Open(connectionsJSONPath())
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

	topUI = newMainUI(configConnections)
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
