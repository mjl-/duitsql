package main

import (
	"fmt"
	"image"
	"strconv"

	"github.com/mjl-/duit"
)

type settingsUI struct {
	duit.Box
}

func newSettingsUI(cc configConnection, done func()) (ui *settingsUI) {
	ui = &settingsUI{}

	origName := cc.Name

	port := ""
	if origName != "" {
		port = fmt.Sprintf("%d", cc.Port)
	}
	var primary *duit.Button
	var conn = struct {
		typePostgres, typeMysql, typeSqlserver     *duit.Radiobutton
		name, host, port, user, password, database *duit.Field
		tls                                        *duit.Checkbox
	}{
		&duit.Radiobutton{Value: "postgres"},
		&duit.Radiobutton{Value: "mysql"},
		&duit.Radiobutton{Value: "sqlserver"},
		&duit.Field{Placeholder: "name...", Text: cc.Name},
		&duit.Field{Placeholder: "localhost", Text: cc.Host},
		&duit.Field{Placeholder: "port...", Text: port},
		&duit.Field{Placeholder: "user name...", Text: cc.User},
		&duit.Field{Placeholder: "password...", Password: true, Text: cc.Password},
		&duit.Field{Placeholder: "database (optional)", Text: cc.Database},
		&duit.Checkbox{Checked: cc.TLS},
	}
	dbtypes := []*duit.Radiobutton{
		conn.typePostgres,
		conn.typeMysql,
		conn.typeSqlserver,
	}
	conn.typePostgres.Group = dbtypes
	conn.typeMysql.Group = dbtypes
	conn.typeSqlserver.Group = dbtypes
	switch cc.Type {
	case "postgres":
		conn.typePostgres.Selected = true
	case "mysql":
		conn.typeMysql.Selected = true
	case "sqlserver":
		conn.typeSqlserver.Selected = true
	}
	radiobuttonValue := func(group []*duit.Radiobutton) interface{} {
		for _, e := range group {
			if e.Selected {
				return e.Value
			}
		}
		return nil
	}

	validPort := func(s string) bool {
		v, err := strconv.ParseInt(s, 10, 32)
		return err == nil && v > 0 && v < 64*1024
	}
	check := func(_ string, r *duit.Event) {
		o := primary.Disabled
		primary.Disabled = conn.name.Text == "" || conn.host.Text == "" || (conn.port.Text != "" && !validPort(conn.port.Text))
		if o != primary.Disabled {
			dui.MarkDraw(primary)
		}
	}
	conn.name.Changed = check
	conn.host.Changed = check
	conn.port.Changed = check

	title := "edit connection"
	action := "save"
	if origName == "" {
		title = "new connection"
		action = "create"
	}
	primary = &duit.Button{
		Text:     action,
		Colorset: &dui.Primary,
		Click: func(r *duit.Event) {
			port := int64(0)
			if conn.port.Text != "" {
				port, _ = strconv.ParseInt(conn.port.Text, 10, 16)
			}
			cc := configConnection{
				Type:     radiobuttonValue(dbtypes).(string),
				Name:     conn.name.Text,
				Host:     conn.host.Text,
				Port:     int(port),
				User:     conn.user.Text,
				Password: conn.password.Text,
				Database: conn.database.Text,
			}
			if origName == "" {
				cUI := newConnUI(cc)
				lv := &duit.ListValue{
					Text:     cc.Name,
					Value:    cUI,
					Selected: true,
				}
				connections.Unselect(nil)
				connections.Values = append([]*duit.ListValue{lv}, connections.Values...)
				connectionBox.Kids = duit.NewKids(cUI)
				dui.MarkDraw(connections)
				dui.MarkLayout(connectionBox)
			} else {
				index := connections.Selected()[0]
				lv := connections.Values[index]
				lv.Value.(*connUI).cc = cc
				lv.Text = cc.Name
			}

			l := make([]configConnection, len(connections.Values)-1)
			for i, lv := range connections.Values[:len(connections.Values)-1] {
				l[i] = lv.Value.(*connUI).cc
			}
			go saveConfigConnections(l)

			done()
		},
	}
	check("", &duit.Event{})
	actionBox := &duit.Box{
		Margin: image.Pt(6, 0),
		Kids:   duit.NewKids(primary),
	}
	if origName != "" {
		cancel := &duit.Button{
			Text: "cancel",
			Click: func(r *duit.Event) {
				done()
			},
		}
		deleteButton := &duit.Button{
			Text:     "delete",
			Colorset: &dui.Danger,
			Click: func(r *duit.Event) {
				defer dui.MarkLayout(nil)

				sel := connections.Selected()
				connections.Values[sel[0]].Selected = false
				connections.Changed(sel[0], &duit.Event{}) // deselects connection

				l := []configConnection{}
				nvalues := []*duit.ListValue{}
				for _, lv := range connections.Values {
					if lv.Value == nil {
						nvalues = append(nvalues, lv)
						continue
					}
					cc := lv.Value.(*connUI).cc
					if cc.Name != origName {
						nvalues = append(nvalues, lv)
						l = append(l, cc)
					}
				}
				connections.Values = nvalues
				go saveConfigConnections(l)
			},
		}
		actionBox.Kids = duit.NewKids(primary, cancel, deleteButton)
	}

	ui.Box.Kids = duit.NewKids(
		duit.NewMiddle(
			&duit.Box{
				MaxWidth: 350,
				Kids: duit.NewKids(
					duit.CenterUI(duit.SpaceXY(4, 2), &duit.Label{Text: title, Font: bold}),
					&duit.Grid{
						Columns: 2,
						Padding: []duit.Space{
							duit.SpaceXY(4, 2),
							duit.SpaceXY(4, 2),
						},
						Halign: []duit.Halign{
							duit.HalignRight,
							duit.HalignLeft,
						},
						Valign: []duit.Valign{
							duit.ValignMiddle,
							duit.ValignMiddle,
						},
						Kids: duit.NewKids(
							label("type"),
							&duit.Box{
								Margin: image.Pt(2, 0),
								Kids: duit.NewKids(
									conn.typePostgres,
									&duit.Label{
										Text: "postgres",
										Click: func(e *duit.Event) {
											conn.typePostgres.Select(dui)
										},
									},
									conn.typeMysql,
									&duit.Label{
										Text: "mysql",
										Click: func(e *duit.Event) {
											conn.typeMysql.Select(dui)
										},
									},
									conn.typeSqlserver,
									&duit.Label{
										Text: "sqlserver",
										Click: func(e *duit.Event) {
											conn.typeSqlserver.Select(dui)
										},
									},
								),
							},
							label("name"),
							conn.name,
							label("host"),
							conn.host,
							label("port"),
							conn.port,
							label("user"),
							conn.user,
							label("password"),
							conn.password,
							label("database"),
							conn.database,
							conn.tls,
							label("require TLS"),
							label(""),
							actionBox,
						),
					},
				),
			},
		),
	)
	return
}
