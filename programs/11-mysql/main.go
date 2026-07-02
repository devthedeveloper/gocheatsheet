// mysql: connect, create, insert, read back.
// needs: go get github.com/go-sql-driver/mysql
// and a server:
//   docker run -d --name mysql-demo \
//     -e MYSQL_ROOT_PASSWORD=secret \
//     -e MYSQL_DATABASE=notes \
//     -p 3306:3306 mysql:8
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

	if err := db.Ping(); err != nil { // real check
		panic("unreachable: " + err.Error())
	}
	fmt.Println("connected ✓")

	db.Exec(`CREATE TABLE IF NOT EXISTS pets (
		id   INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(50) NOT NULL)`)

	res, err := db.Exec(
		"INSERT INTO pets (name) VALUES (?)",
		"Gopher")
	if err != nil {
		panic(err)
	}
	id, _ := res.LastInsertId()
	fmt.Println("inserted pet with id", id)

	var name string
	db.QueryRow("SELECT name FROM pets WHERE id = ?",
		id).Scan(&name)
	fmt.Println("read back:", name)
}
