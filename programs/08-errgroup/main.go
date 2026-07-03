// errgroup: POST /checkout fans out to 3 real DB updates; payment declines.
// needs: go get golang.org/x/sync/errgroup
package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/sync/errgroup"
)

func setup(db *sql.DB) {
	db.Exec("DROP TABLE IF EXISTS inventory, wallets, shipments")
	db.Exec(`CREATE TABLE inventory (sku VARCHAR(20) PRIMARY KEY, stock INT)`)
	db.Exec(`CREATE TABLE wallets (user_id INT PRIMARY KEY, balance_paisa INT)`)
	db.Exec(`CREATE TABLE shipments (order_id INT PRIMARY KEY, sku VARCHAR(20))`)
	db.Exec("INSERT INTO inventory VALUES ('boat-airdopes', 40)")
	db.Exec("INSERT INTO wallets VALUES (7, 5000)") // ₹50 — not enough
}

func reserveStock(db *sql.DB, sku string, qty int) error {
	res, err := db.Exec(
		"UPDATE inventory SET stock = stock - ? WHERE sku=? AND stock >= ?",
		qty, sku, qty)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("inventory-service: out of stock")
	}
	fmt.Println("inventory-service → reserved", qty, sku)
	return nil
}

func chargeWallet(db *sql.DB, userID, amountPaisa int) error {
	// a REAL conditional UPDATE: only succeeds if balance covers it
	res, err := db.Exec(
		"UPDATE wallets SET balance_paisa = balance_paisa - ? "+
			"WHERE user_id=? AND balance_paisa >= ?",
		amountPaisa, userID, amountPaisa)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("payment-service: insufficient balance")
	}
	fmt.Println("payment-service → charged ₹", amountPaisa/100)
	return nil
}

func createShipment(db *sql.DB, orderID int, sku string) error {
	_, err := db.Exec(
		"INSERT INTO shipments VALUES (?, ?)", orderID, sku)
	if err == nil {
		fmt.Println("shipping-service → shipment created")
	}
	return err
}

func main() {
	db, err := sql.Open("mysql",
		"root:secret@tcp(localhost:3306)/notes")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	setup(db)

	var g errgroup.Group
	g.Go(func() error { return reserveStock(db, "boat-airdopes", 1) })
	g.Go(func() error { return chargeWallet(db, 7, 249900) }) // ₹2499, wallet has ₹50
	g.Go(func() error { return createShipment(db, 1042, "boat-airdopes") })

	if err := g.Wait(); err != nil {
		fmt.Println("checkout failed:", err)
		fmt.Println("→ respond 402, release the reserved stock")
		db.Exec("UPDATE inventory SET stock = stock + 1 WHERE sku='boat-airdopes'")
	}
}
