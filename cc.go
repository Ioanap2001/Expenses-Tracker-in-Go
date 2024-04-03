package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite-Treiber
)



// Definition der Transaction-Struktur
type Transaction struct {
	ID          int
	Amount      float64
	Description string
	Category    string
	Date        string
}

func main() {
	var err error

	// Datenbankverbindung herstellen
	db, err = sql.Open("sqlite3", "ear.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTables()

	fmt.Println("EAR - Kreditkartenabrechnung gestartet!")

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/addTransaction", handleAddTransaction)
	http.HandleFunc("/viewTransactions", handleViewTransactions)
	http.HandleFunc("/addIncome", handleAddIncome)
	http.HandleFunc("/calculateExpenses", handleCalculateExpenses)
	http.HandleFunc("/exit", handleExit)
	http.HandleFunc("/calculatePayment", handleCalculatePayment)
	http.HandleFunc("/export", handleExport)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", nil)
}

func handleAddTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, "addTransaction", nil)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	amount := r.FormValue("amount")
	description := r.FormValue("description")
	category := r.FormValue("category")
	date := r.FormValue("date")

	// Datenbank einfügen
	_, err := db.Exec("INSERT INTO credit_card_transactions(amount, description, category, date) VALUES (?, ?, ?, ?)", amount, description, category, date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Kreditkartentransaktion erfolgreich hinzugefügt!")
}

func handleViewTransactions(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM credit_card_transactions ORDER BY date DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []Transaction

	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.Amount, &t.Description, &t.Category, &t.Date); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "viewTransactions", transactions)
}

func handleAddIncome(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, "addIncome", nil)
		return
	}

}

func handleCalculateExpenses(w http.ResponseWriter, r *http.Request) {
	// Query the database to get monthly expenses
	rows, err := db.Query("SELECT category, SUM(amount) AS total FROM credit_card_transactions GROUP BY category")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Define a struct to hold the data to be displayed in the template
	type Expense struct {
		Category string
		Total    float64
	}
	var expenses []Expense

	// Iterate through the rows and populate the expenses slice
	var totalExpenses float64
	for rows.Next() {
		var expense Expense
		if err := rows.Scan(&expense.Category, &expense.Total); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		expenses = append(expenses, expense)
		totalExpenses += expense.Total
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Pass the expenses data to the template
	data := struct {
		Expenses      []Expense
		TotalExpenses float64
	}{
		Expenses:      expenses,
		TotalExpenses: totalExpenses,
	}
	renderTemplate(w, "calculateExpenses", data)
}

func handleExit(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "exit", nil)
}

func handleCalculatePayment(w http.ResponseWriter, r *http.Request) {
	// Annahme: Mindestzahlung ist 10% der gesamten Kreditkartenausgaben in diesem Monat
	rows, err := db.Query("SELECT SUM(amount) FROM credit_card_transactions WHERE strftime('%Y-%m', date) = strftime('%Y-%m', 'now')")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var totalExpenses float64
	rows.Next()
	if err := rows.Scan(&totalExpenses); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	minimumPayment := totalExpenses * 0.1 // Mindestzahlung ist 10% der gesamten Ausgaben

	// Annahme: Fälligkeitsdatum ist 25 Tage nach Ende des aktuellen Monats
	currentDate := time.Now()
	lastDayOfMonth := time.Date(currentDate.Year(), currentDate.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	dueDate := lastDayOfMonth.AddDate(0, 0, 25) // Fälligkeitsdatum ist 25 Tage nach Ende des Monats

	daysUntilDue := int(dueDate.Sub(currentDate).Hours() / 24)

	// Render the template with payment data
	data := struct {
		MinimumPayment float64
		DueDate        string
		DaysUntilDue   int
	}{
		MinimumPayment: minimumPayment,
		DueDate:        dueDate.Format("2006-01-02"),
		DaysUntilDue:   daysUntilDue,
	}
	renderTemplate(w, "calculatePayment", data)
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	// Setze den Content-Type auf CSV
	w.Header().Set("Content-Type", "text/csv")
	// Setze den Content-Disposition Header, um den Browser zum Herunterladen der Datei zu veranlassen
	w.Header().Set("Content-Disposition", "attachment; filename=ausgaben.csv")

	// Ausgaben in CSV schreiben
	exportSummaryToCSV(w)
}

func handleExpensesData(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT category, SUM(amount) FROM credit_card_transactions GROUP BY category")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	data := make(map[string]float64)
	for rows.Next() {
		var category string
		var amount float64
		if err := rows.Scan(&category, &amount); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data[category] = amount
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the data as JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}


	// Send the data as JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}



	// Ausgaben aus der Datenbank abrufen
	rows, err := db.Query("SELECT category, SUM(amount) FROM credit_card_transactions GROUP BY category")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// CSV-Datei schreiben
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// CSV-Header schreiben
	csvWriter.Write([]string{"Kategorie", "Gesamtausgaben"})

	// Daten aus der Abfrage in die CSV-Datei schreiben
	for rows.Next() {
		var category string
		var totalAmount float64

		if err := rows.Scan(&category, &totalAmount); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		

		if err := csvWriter.Write([]string{category, fmt.Sprintf("%.2f", totalAmount)}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Println("Die Zusammenfassung wurde erfolgreich in CSV exportiert.")
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles("templates/" + tmpl + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createTables() {
	_, err := db.Exec(`
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
