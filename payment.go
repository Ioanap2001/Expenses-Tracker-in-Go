package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite-Treiber
)

func calculatePaymentDueDate() {
	// Datenbankverbindung herstellen
	db, err := sql.Open("sqlite3", "ear.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Annahme: Mindestzahlung ist 10% der gesamten Kreditkartenausgaben in diesem Monat
	rows, err := db.Query("SELECT SUM(amount) FROM credit_card_transactions WHERE strftime('%Y-%m', date) = strftime('%Y-%m', 'now')")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var totalExpenses float64
	rows.Next()
	err = rows.Scan(&totalExpenses)
	if err != nil {
		log.Fatal(err)
	}

	minimumPayment := totalExpenses * 0.1 // Mindestzahlung ist 10% der gesamten Ausgaben

	// Annahme: Fälligkeitsdatum ist 25 Tage nach Ende des aktuellen Monats
	currentDate := time.Now()
	lastDayOfMonth := time.Date(currentDate.Year(), currentDate.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	dueDate := lastDayOfMonth.AddDate(0, 0, 25) // Fälligkeitsdatum ist 25 Tage nach Ende des Monats

	daysUntilDue := int(dueDate.Sub(currentDate).Hours() / 24)

	fmt.Printf("Mindestzahlung: €%.2f\n", minimumPayment)
	fmt.Printf("Fälligkeitsdatum: %s\n", dueDate.Format("2006-01-02"))
	fmt.Printf("%d Tage bis zur nächsten Zahlung.\n", daysUntilDue)
}
