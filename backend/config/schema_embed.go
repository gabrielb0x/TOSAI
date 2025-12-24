package configdata

import "embed"

// SchemaFS embarque le DDL PostgreSQL pour initialiser le schéma TOSAI.
//
//go:embed database_init.sql
var SchemaFS embed.FS
