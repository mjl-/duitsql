package main

import (
	"fmt"
	"image"
	"strconv"

	"github.com/mjl-/duit"
)

type settingsUI struct {
	*duit.Box
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
		name, host, port, user, password, database *duit.Field
		tls                                        *duit.Checkbox
	}{
		&duit.Field{Placeholder: "name...", Text: cc.Name},
		&duit.Field{Placeholder: "localhost", Text: cc.Host},
		&duit.Field{Placeholder: "5432", Text: port},
		&duit.Field{Placeholder: "user name...", Text: cc.User},
		&duit.Field{Placeholder: "password...", Password: true, Text: cc.Password},
		&duit.Field{Placeholder: "database (optional)", Text: cc.Database},
		&duit.Checkbox{Checked: cc.TLS},
	}
	validPort := func(s string) bool {
		v, err := strconv.ParseInt(s, 10, 32)
		return err == nil && v > 0 && v < 64*1024
	}
	check := func(_ string, r *duit.Result) {
		primary.Disabled = conn.name.Text == "" || conn.host.Text == "" || (conn.port.Text != "" && !validPort(conn.port.Text))
		r.Draw = true
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
		Text:    action,
		Primary: true,
		Click: func(r *duit.Result) {
			port := int64(5432)
			if conn.port.Text != "" {
				port, _ = strconv.ParseInt(conn.port.Text, 10, 16)
			}
			cc := configConnection{
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
			} else {
				index := connections.Selected()[0]
				lv := connections.Values[index]
				lv.Value.(*connUI).cc = cc
				lv.Text = cc.Name
			}

			r.Layout = true

			l := make([]configConnection, len(connections.Values)-1)
			for i, lv := range connections.Values[:len(connections.Values)-1] {
				l[i] = lv.Value.(*connUI).cc
			}
			go saveConfigConnections(l)

			done()
		},
	}
	check("", &duit.Result{})
	actionBox := &duit.Box{
		Margin: image.Pt(6, 0),
		Kids:   duit.NewKids(primary),
	}
	if origName != "" {
		cancel := &duit.Button{
			Text: "cancel",
			Click: func(r *duit.Result) {
				done()
			},
		}
		deleteButton := &duit.Button{
			Text: "delete",
			Click: func(r *duit.Result) {
				sel := connections.Selected()
				connections.Values[sel[0]].Selected = false
				connections.Changed(sel[0], &duit.Result{}) // deselects connection

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
				r.Layout = true
				go saveConfigConnections(l)
			},
		}
		actionBox.Kids = duit.NewKids(primary, cancel, deleteButton)
	}

	ui.Box = &duit.Box{
		Kids: duit.NewKids(
			duit.NewMiddle(
				&duit.Box{
					MaxWidth: 350,
					Kids: duit.NewKids(
						duit.CenterUI(&duit.Label{Text: title, Font: bold}, duit.SpaceXY(4, 2)),
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
								&duit.Label{Text: "name"},
								conn.name,
								&duit.Label{Text: "host"},
								conn.host,
								&duit.Label{Text: "port"},
								conn.port,
								&duit.Label{Text: "user"},
								conn.user,
								&duit.Label{Text: "password"},
								conn.password,
								&duit.Label{Text: "database"},
								conn.database,
								conn.tls,
								&duit.Label{Text: "require TLS"},
								&duit.Label{},
								actionBox,
							),
						},
					),
				},
			),
		),
	}
	return
}
