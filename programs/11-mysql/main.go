// mysql: what POST /signup does to the database.
// needs: go get github.com/go-sql-driver/mysql
// and a server:
//   docker run -d --name mysql-demo \
//     -e MYSQL_ROOT_PASSWORD=secret \
//     -e MYSQL_DATABASE=notes -p 3306:3306 mysql:8
package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql",
		"root:secret@tcp(localhost:3306)/notes")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic("db unreachable: " + err.Error())
	}
	fmt.Println("connected ✓")

	db.Exec("DROP TABLE IF EXISTS users") // rerunnable demo
	db.Exec(`CREATE TABLE users (
		id    INT AUTO_INCREMENT PRIMARY KEY,
		email VARCHAR(120) NOT NULL UNIQUE)`)

	// the signup form's email goes in via ? — never
	// by pasting it into the SQL string
	res, err := db.Exec(
		"INSERT INTO users (email) VALUES (?)",
		"asha@example.com")
	if err != nil {
		panic(err) // e.g. duplicate email → 409
	}
	id, _ := res.LastInsertId()
	fmt.Println("new user id:", id)

	var email string
	db.QueryRow(
		"SELECT email FROM users WHERE id = ?",
		id).Scan(&email)
	fmt.Println("stored:", email)
}
