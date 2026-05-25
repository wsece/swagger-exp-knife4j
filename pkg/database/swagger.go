// swagger.go: Opens a GORM connection by URI scheme and migrates the SwaggerAPIRecord table.
package database

import (
	"errors"
	"net/url"
	"strings"

	"swagger-exp-knife4j/pkg/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SwaggerConnection opens a database (sqlite/postgres/mysql) by uri and AutoMigrates SwaggerAPIRecord.
// uri is the connection string; when debug is true, GORM SQL is logged. Returns *gorm.DB or a connect/migrate error.
func SwaggerConnection(uri string, debug bool) (*gorm.DB, error) {
	dbURL, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	config := &gorm.Config{}
	if debug {
		config.Logger = logger.Default.LogMode(logger.Info)
	} else {
		config.Logger = logger.Default.LogMode(logger.Error)
	}

	var conn *gorm.DB
	switch dbURL.Scheme {
	case "sqlite":
		dbPath := sqliteDBPath(dbURL)
		conn, err = gorm.Open(sqlite.Open(dbPath+"?cache=shared"), config)
		if err != nil {
			return nil, err
		}
		conn.Exec("PRAGMA foreign_keys = ON")
	case "postgres":
		conn, err = gorm.Open(postgres.Open(uri), config)
		if err != nil {
			return nil, err
		}
	case "mysql":
		conn, err = gorm.Open(mysql.Open(uri), config)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid db uri scheme")
	}

	if err := conn.AutoMigrate(&models.SwaggerAPIRecord{}); err != nil {
		return nil, err
	}

	return conn, nil
}

func sqliteDBPath(dbURL *url.URL) string {
	if dbURL.Path != "" && dbURL.Path != "/" {
		p := dbURL.Path
		if strings.HasPrefix(p, "/") {
			p = p[1:]
		}
		return p
	}
	return dbURL.Host
}
