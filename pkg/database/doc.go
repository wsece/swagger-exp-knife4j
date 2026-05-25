// Package database provides scan-result database connections and schema migration (infrastructure layer).
//
// Supported URI schemes: sqlite, postgres, mysql.
// SwaggerConnection opens a connection and runs AutoMigrate on models.SwaggerAPIRecord.
package database
