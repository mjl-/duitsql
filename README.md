duitsql - simple sql database browser and query executor

# about

Duitsql lets you connect to SQL database (only postgres currently), and list databases, tables & views, and data in them. It also lets you execute queries and view the results.
Duitsql was created to showcase duit, the developer ui toolkit, and vice versa.

# license

MIT-license, see LICENSE.md

# todo

- improve showing structure for tables, try to keep it cross-database through information_schema.  should show check constraints, foreign key constraint, indexes.
- new setting by duplicating existing one

- fetch rows from resultset on demand. requires updating duit.Gridlist. and/or do paging.

- add buttons to refresh list of database, list of tables/views, data for table/view
- fix todo's
- search for connections, databases, tables, at bottom of list

- find cause of not being able to cancel queries that cause a new db connection to be created
- colors in UI to indicate which connection you have open
