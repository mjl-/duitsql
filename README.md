duitsql - simple sql database browser and query executor

# about

Duitsql lets you connect to SQL database (only postgres currently), and list databases, tables & views, and data in them. It also lets you execute queries and view the results.
Duitsql was created to showcase duit, the developer ui toolkit, and vice versa.

# license

MIT-license, see LICENSE.md

# todo

- on invalid credentials (eg bad password/username) while connecting, no error message is shown, but it is still "connecting..."
- add buttons to refresh list of database, list of tables/views, data for table/view
- show difference between table and view
- show time it is taking to run query
- allow aborting query
- show structure of tables (different from self-written queries). find queries for postgres. wrap them in an interface, for easier future other database support.
- fix todo's, like disconnecting properly
- fetch rows from resultset on demand. requires updating duit.Gridlist
- search for connections, databases, tables, at bottom of list

- show []byte as hex?

- new setting by duplicating existing one
- connect via ssh?
- understand postgis responses, and show geo data (if small enough)
- support more databases: mysql, mssql
- colors in UI to indicate which connection you have open
