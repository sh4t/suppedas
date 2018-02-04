package main

import (
	"database/sql"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func databaseWriter(dbFile string, persistChannel chan persistMessage) {

	db, err := sql.Open("sqlite3", dbFile)
	check(err)
	defer db.Close()

	createTables(db)

	rssiStmt, err := db.Prepare("INSERT INTO rssi(time, mac, rssi) values(?,?,?)")
	check(err)
	nameStmt, err := db.Prepare("INSERT INTO name(time, mac, name) values(?,?,?)")
	check(err)

	tx, err := db.Begin()
	check(err)
	// set low while developing
	commitIntervalSeconds := 1.0
	lastCommit := time.Now()
	for {
		entry := <-persistChannel
		// if rssi is empty this is a name entry
		if entry.Name == "" {
			_, err := rssiStmt.Exec(entry.Timestamp, entry.Mac, entry.Rssi)
			check(err)
		} else {
			_, err := nameStmt.Exec(entry.Timestamp, entry.Mac, entry.Name)
			check(err)
		}
		if time.Since(lastCommit).Seconds() > commitIntervalSeconds {
			tx.Commit()
			lastCommit = time.Now()
		}
	}
}

func createTables(db *sql.DB) {
	sqlStmt := `
	create table rssi (id integer not null primary key, time text, mac text, rssi text);
	create table name (id integer not null primary key, time text, mac text, name text);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			log.Printf("%q: %s\n", err, sqlStmt)
			check(err)
		}
	}

}
