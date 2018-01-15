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

func newSettingsUI(cc configConnection, isNew bool, done func()) (ui *settingsUI) {
	ui = &settingsUI{}

	origName := cc.Name

	port := ""
	if cc.Port != 0 {
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
	check := func(_ string) (e duit.Event) {
		o := primary.Disabled
		primary.Disabled = conn.name.Text == "" || conn.host.Text == "" || (conn.port.Text != "" && !validPort(conn.port.Text))
		if o != primary.Disabled {
			dui.MarkDraw(primary)
		}
		return
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
		Click: func() (e duit.Event) {
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
			l := make([]configConnection, len(topUI.connectionList.Values)-1)
			for i, lv := range topUI.connectionList.Values[:len(topUI.connectionList.Values)-1] {
				if lv.Selected {
					l[i] = cc
				} else {
					l[i] = lv.Value.(*connUI).cc
				}
			}

			go func() {
				saveConfigConnections(l)

				dui.Call <- func() {
					if origName == "" {
						topUI.newConnection(cc)
					} else {
						topUI.updateConnection(cc)
					}
					done()
				}
			}()
			return
		},
	}
	check("")
	actionBox := &duit.Box{
		Margin: image.Pt(6, 0),
		Kids:   duit.NewKids(primary),
	}
	if origName != "" {
		cancel := &duit.Button{
			Text: "cancel",
			Click: func() (e duit.Event) {
				done()
				return
			},
		}
		deleteButton := &duit.Button{
			Text:     "delete",
			Colorset: &dui.Danger,
			Click: func() (e duit.Event) {
				topUI.deleteSelectedConnection()
				return
			},
		}
		buttons := []duit.UI{primary, cancel, deleteButton}
		if !isNew {
			duplicate := &duit.Button{
				Text: "duplicate",
				Click: func() (e duit.Event) {
					ncc := cc
					ncc.Name = ""
					topUI.duplicateSettings(ncc)
					return
				},
			}
			buttons = append(buttons, duplicate)
		}
		actionBox.Kids = duit.NewKids(buttons...)
	}

	ui.Box.Kids = duit.NewKids(
		duit.NewMiddle(
			duit.SpaceXY(10, 10),
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
										Click: func() (e duit.Event) {
											conn.typePostgres.Select(dui)
											return
										},
									},
									conn.typeMysql,
									&duit.Label{
										Text: "mysql",
										Click: func() (e duit.Event) {
											conn.typeMysql.Select(dui)
											return
										},
									},
									conn.typeSqlserver,
									&duit.Label{
										Text: "sqlserver",
										Click: func() (e duit.Event) {
											conn.typeSqlserver.Select(dui)
											return
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
