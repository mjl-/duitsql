package main

import (
	"fmt"
	"image"
	"strconv"

	"github.com/mjl-/duit"
)

type settingsUI struct {
	typePostgres, typeMysql, typeSqlserver     *duit.Radiobutton
	name, host, port, user, password, database *duit.Field
	tls                                        *duit.Checkbox

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
	ui.typePostgres = &duit.Radiobutton{Value: "postgres"}
	ui.typeMysql = &duit.Radiobutton{Value: "mysql"}
	ui.typeSqlserver = &duit.Radiobutton{Value: "sqlserver"}
	ui.name = &duit.Field{Placeholder: "name...", Text: cc.Name}
	ui.host = &duit.Field{Placeholder: "localhost", Text: cc.Host}
	ui.port = &duit.Field{Placeholder: "port...", Text: port}
	ui.user = &duit.Field{Placeholder: "user name...", Text: cc.User}
	ui.password = &duit.Field{Placeholder: "password...", Password: true, Text: cc.Password}
	ui.database = &duit.Field{Placeholder: "database (optional)", Text: cc.Database}
	ui.tls = &duit.Checkbox{Checked: cc.TLS}

	dbTypes := []*duit.Radiobutton{
		ui.typePostgres,
		ui.typeMysql,
		ui.typeSqlserver,
	}
	ui.typePostgres.Group = dbTypes
	ui.typeMysql.Group = dbTypes
	ui.typeSqlserver.Group = dbTypes
	switch cc.Type {
	case "postgres":
		ui.typePostgres.Selected = true
	case "mysql":
		ui.typeMysql.Selected = true
	case "sqlserver":
		ui.typeSqlserver.Selected = true
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
		primary.Disabled = ui.name.Text == "" || ui.host.Text == "" || (ui.port.Text != "" && !validPort(ui.port.Text))
		if o != primary.Disabled {
			dui.MarkDraw(primary)
		}
		return
	}
	ui.name.Changed = check
	ui.host.Changed = check
	ui.port.Changed = check

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
			if ui.port.Text != "" {
				port, _ = strconv.ParseInt(ui.port.Text, 10, 16)
			}
			cc := configConnection{
				Type:     radiobuttonValue(dbTypes).(string),
				Name:     ui.name.Text,
				Host:     ui.host.Text,
				Port:     int(port),
				User:     ui.user.Text,
				Password: ui.password.Text,
				Database: ui.database.Text,
			}
			l := make([]configConnection, len(topUI.connectionList.Values)-1)
			for i, lv := range topUI.connectionList.Values[:len(topUI.connectionList.Values)-1] {
				if lv.Selected {
					l[i] = cc
				} else {
					l[i] = lv.Value.(*connUI).cc
				}
			}

			if origName == "" {
				topUI.addNewConnection(cc)
			} else {
				topUI.updateSelectedConnection(cc)
			}
			done()

			topUI.saveConnections()
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
				topUI.saveConnections()
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
									ui.typePostgres,
									&duit.Label{
										Text: "postgres",
										Click: func() (e duit.Event) {
											ui.typePostgres.Select(dui)
											return
										},
									},
									ui.typeMysql,
									&duit.Label{
										Text: "mysql",
										Click: func() (e duit.Event) {
											ui.typeMysql.Select(dui)
											return
										},
									},
									ui.typeSqlserver,
									&duit.Label{
										Text: "sqlserver",
										Click: func() (e duit.Event) {
											ui.typeSqlserver.Select(dui)
											return
										},
									},
								),
							},
							label("name"),
							ui.name,
							label("host"),
							ui.host,
							label("port"),
							ui.port,
							label("user"),
							ui.user,
							label("password"),
							ui.password,
							label("database"),
							ui.database,
							ui.tls,
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
