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
	dui, err = duit.NewDUI("sql", nil)
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

	topUI = newMainUI(configConnections)
	dui.Top.UI = topUI
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
