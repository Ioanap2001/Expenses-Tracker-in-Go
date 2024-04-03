package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3" // SQLite-Treiber
)

func exportSummaryToCSV() {
	// Datenbankverbindung herstellen
	db, err := sql.Open("sqlite3", "ear.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Ausgaben aus der Datenbank abrufen
	rows, err := db.Query("SELECT category, SUM(amount) FROM credit_card_transactions GROUP BY category")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// CSV-Datei schreiben
	filename := getInput("Geben Sie den Dateinamen für die CSV-Datei ein (ohne Erweiterung): ") + ".csv"
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	// CSV-Header schreiben
	csvWriter.Write([]string{"Kategorie", "Gesamtausgaben"})

	// Daten aus der Abfrage in die CSV-Datei schreiben
	for rows.Next() {
		var category string
		var totalAmount float64

		err := rows.Scan(&category, &totalAmount)
		if err != nil {
			log.Fatal(err)
		}

		err = csvWriter.Write([]string{category, fmt.Sprintf("%.2f", totalAmount)})
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("Die Zusammenfassung wurde erfolgreich in '%s' exportiert.\n", filename)
}

func getInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func main() {
	exportSummaryToCSV()
}
