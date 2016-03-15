package main

import (
	"fmt"
	"os"
)

func createTables() {
	stmt, _ := db.Prepare(`CREATE TABLE messages (
		id serial PRIMARY KEY,
		username varchar NOT NULL,
		message varchar,
		created_at timestamp not null default now()
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

	stmt.Close()
}

func insertMessage(m message) {
	message := m.Message
	username := m.Username

	insert, err := db.Prepare("INSERT INTO messages (username, message, created_at) VALUES ($1, $2, $3)")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't preapre statement")
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer insert.Close()

	tx, err := db.Begin()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't begin transaction")
		fmt.Fprintln(os.Stderr, err)
		return
	}

	_, err = tx.Stmt(insert).Exec(username, message, m.Timestamp)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Couldn't execute statement")
		fmt.Fprintln(os.Stderr, err)
		tx.Rollback()
		return
	}

	tx.Commit()
}

func recentMessages() []message {
	rows, err := db.Query("SELECT username, message, created_at FROM messages ORDER BY id DESC LIMIT 50")
	if err != nil {
		panic(err)
	}

	messages := []message{}
	for rows.Next() {
		var m message
		rows.Scan(&m.Username, &m.Message, &m.Timestamp)

		messages = append(messages, m)
	}

	rows.Close()

	return messages
}
