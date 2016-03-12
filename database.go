package main

import (
	"fmt"
	"os"
)

func createTables() {
	stmt, _ := db.Prepare(`CREATE TABLE messages (
		id serial,
		username varchar,
		message varchar
	)`)

	tx, _ := db.Begin()
	_, err := tx.Stmt(stmt).Exec()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		tx.Rollback()
		return
	}

	tx.Commit()
}

func insertMessage(username, message string) {
	insert, err := db.Prepare("INSERT INTO messages (username, message) VALUES (?, ?)")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	_, err = tx.Stmt(insert).Exec(username, message)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		tx.Rollback()
		return
	}

	tx.Commit()
}
