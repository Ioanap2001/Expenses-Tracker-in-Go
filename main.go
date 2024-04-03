package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite-Treiber
)

var db *sql.DB

func main() {
	var err error

	// Datenbankverbindung herstellen
	db, err = sql.Open("sqlite3", "ear.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Tabellen erstellen, falls sie nicht existieren
	createTables()

	fmt.Println("Einnahme-Ausgabe-Rechnung gestartet!")

	// Benutzereingabe verarbeiten
	for {
		printMenu()
		processInput()
	}
}

func createTables() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS expenses(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			amount REAL,
			category TEXT,
			description TEXT,
			date DATE
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS credit_card_transactions(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			amount REAL,
			description TEXT,
			category TEXT, 
			date DATE
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func printMenu() {
	fmt.Println("\nEAR")
	fmt.Println("1. Transaktion hinzufügen")
	fmt.Println("2. Verlauf anschauen")
	fmt.Println("3. Verlassen")
}

func processInput() {
	var choice string
	fmt.Print("Auswählen: ")
	_, err := fmt.Scanln(&choice)
	if err != nil {
		log.Fatal(err)
	}

	switch choice {
	case "1":
		addTransaction()
	case "2":
		viewExpenses()
	case "3":
		fmt.Println("Verlassen...\n")
		os.Exit(0)
	default:
		fmt.Println("Ungültige Auswahl.")
	}
}

func addTransaction() {
	var amount float64
	var category, description string

	fmt.Print("Betrag: ")
	_, err := fmt.Scanln(&amount)
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Kategorie: ")
	category, err = reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Beschreibung: ")
	description, err = reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	// Datenbank einfügen
	date := time.Now().Format("2006-01-02")
	_, err = db.Exec("INSERT INTO expenses(amount, category, description, date) VALUES (?, ?, ?, ?)", amount, category, description, date)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaktion erfolgreich hinzugefügt!")
}

func viewExpenses() {
	rows, err := db.Query("SELECT * FROM expenses ORDER BY date")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Nr.   / Betrag  / Kategorie  / Beschreibung    / Datum ")
	fmt.Println("___________________________________________________________")

	for rows.Next() {
		var id int
		var amount float64
		var category, description, date string

		err := rows.Scan(&id, &amount, &category, &description, &date)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%-3d  / €%-7.2f  / %-15s / %-27s  / %s\n", id, amount, category, description, date)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}
