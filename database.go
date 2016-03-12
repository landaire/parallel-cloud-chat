package main

import (
	"fmt"
	"os"
)

func createTables() {
	stmt, _ := db.Prepare(`CREATE TABLE messages (
		id serial PRIMARY KEY,
		username varchar,
		message varchar
	)`)

	tx, err := db.Begin()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	_, err = tx.Stmt(stmt).Exec()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		tx.Rollback()
		return
	}

	tx.Commit()
}

func insertMessage(username, message string) {
	insert, err := db.Prepare("INSERT INTO messages (username, message) VALUES ($1, $2)")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't preapre statement")
		fmt.Fprintln(os.Stderr, err)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't begin transaction")
		fmt.Fprintln(os.Stderr, err)
		return
	}

	_, err = tx.Stmt(insert).Exec(username, message)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't execute statement")
		fmt.Fprintln(os.Stderr, err)
		tx.Rollback()
		return
	}

	tx.Commit()
}
