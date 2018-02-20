package main

import (
	"fmt"
	"net/url"
	"strings"
)

type connectionConfig struct {
	Type     string // postgres, mysql, sqlserver
	Name     string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	TLS      bool
}

// connectionString returns an URL or connection string that can be passed to sql.Open.
func (c connectionConfig) connectionString(dbName string) string {
	switch c.Type {
	case "postgres":
		quote := func(s string) string {
			s = strings.Replace(s, `\`, `\\`, -1)
			s = strings.Replace(s, `'`, `\'`, -1)
			if s == "" || strings.ContainsAny(s, " '\\") {
				s = "'" + s + "'"
			}
			return s
		}
		sslmode := "disable"
		if c.TLS {
			sslmode = "verify-full"
		}
		s := fmt.Sprintf("host=%s sslmode=%s application_name=duitsql", quote(c.Host), sslmode)
		if c.Port != 0 {
			s += fmt.Sprintf(" port=%d", c.Port)
		}
		if c.User != "" {
			s += fmt.Sprintf(" user=%s", quote(c.User))
		}
		if c.Password != "" {
			s += fmt.Sprintf(" password=%s", quote(c.Password))
		}
		if dbName != "" {
			s += fmt.Sprintf(" dbname=%s", quote(dbName))
		}
		return s
	case "mysql":
		s := ""
		if c.User != "" || c.Password != "" {
			s += c.User
			if c.Password != "" {
				s += ":" + c.Password
			}
			s += "@"
		}
		address := c.Host
		if c.Port != 0 {
			address += fmt.Sprintf(":%d", c.Port)
		}
		s += fmt.Sprintf("tcp(%s)", address)
		s += "/"
		if dbName != "" {
			s += dbName
		}
		if c.TLS {
			s += "?tls=true"
		}
		return s
	case "sqlserver":
		host := c.Host
		if c.Port != 0 {
			host += fmt.Sprintf(":%d", c.Port)
		}
		qs := []string{}
		if dbName != "" {
			qs = append(qs, "database="+url.QueryEscape(dbName))
		}
		if c.TLS {
			qs = append(qs, "encrypt=true", "TrustServerCertificate=false")
		}
		u := &url.URL{
			Scheme:   "sqlserver",
			User:     url.UserPassword(c.User, c.Password),
			Host:     host,
			RawQuery: strings.Join(qs, "&"),
		}
		return u.String()
	}
	panic("missing case")
}
