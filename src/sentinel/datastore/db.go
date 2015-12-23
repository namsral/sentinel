// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datastore

import (
	"log"
	"sync"

	// Import the postgres driver
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

var (
	// DB is the db instance
	DB          *sqlx.DB
	connectOnce sync.Once
	createSQL   []string
)

func init() {
	// FIXME: calling Connect() within init() is a quick fix as calling
	// cmd/sentinel.serveCMD:Connnect() results in a nil database.db
	Connect()
}

// Connect connects to the database and asigns the db instance
func Connect() {
	connectOnce.Do(func() {
		var err error
		DB, err = sqlx.Open("postgres", "")
		if err != nil {
			log.Fatal("Error preparing PostgreSQL database instance using environment variables")
		}
		if err = DB.Ping(); err != nil {
			log.Fatal("Error connecting to the database")
		}
	})
}

// Create creates the db tables
func Create() {
	createSQL = []string{
		userTableCreateStmt,
		authemailTableCreateStmt,
		serviceTableCreateStmt,
	}
	for _, query := range createSQL {
		if _, err := DB.Exec(query); err != nil {
			log.Fatalf("Error running query %q: %s", query, err)
		}
	}
}

// Drop drops the db tables
func Drop() {
	// DB.Exec(`DROP INDEX IF EXISTS user_isarchived;`)
	dropTables := []string{
		authemailTable,
		userTable,
		serviceTable,
	}
	for _, t := range dropTables {
		if _, err := DB.Exec(`DROP TABLE IF EXISTS ` + t + `;`); err != nil {
			log.Println("Error dropping table", t, err)
		}
	}
}
