package main

import (
	"fmt"
	"net/url"
	"strings"
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
