// Master server for a distributed database system
package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Column struct {
	Name     string
	DataType string
}

type Table struct {
	Name    string
	Columns []Column
}

func allowCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

var db *sql.DB
var slaveAddresses = []string{"http://192.168.1.4:8002", "http://192.168.1.8:8005", "http://192.168.1.6:8004", "http://192.168.1.7:8006"}
var isMaster bool = true
var masterAddress string = "http://localhost:8001"
var electionInProgress bool = false

func main() {
	var err error
	db, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Define routes
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})
	http.HandleFunc("/nodes", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"master":   masterAddress,
			"slaves":   slaveAddresses,
			"isMaster": isMaster,
		})
	})
	http.HandleFunc("/is-master", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{
			"isMaster": isMaster,
		})
	})

	http.HandleFunc("/createdb", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		createDB(w, r)
	})

	http.HandleFunc("/dropdb", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		dropDB(w, r)
	})

	http.HandleFunc("/createtable", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		createTable(w, r)
	})

	http.HandleFunc("/insert", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		insertRecord(w, r)
	})

	http.HandleFunc("/select", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		selectRecords(w, r)
	})

	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		updateRecord(w, r)
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		deleteRecord(w, r)
	})

	go checkMasterHealth()
	fmt.Println("Master server running on port 8001...")
	log.Fatal(http.ListenAndServe(":8001", nil))
}

func createDB(w http.ResponseWriter, r *http.Request) {
	dbname := r.URL.Query().Get("name")
	if dbname == "" {
		http.Error(w, "Database name is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbname)
	if err != nil {
		http.Error(w, "Failed to create database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	go replicateToSlaves("/replicate/db?name=" + dbname)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Database created successfully"})
}

func dropDB(w http.ResponseWriter, r *http.Request) {
	dbname := r.URL.Query().Get("name")
	if dbname == "" {
		http.Error(w, "Database name is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DROP DATABASE IF EXISTS " + dbname)
	if err != nil {
		http.Error(w, "Failed to drop database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	go replicateToSlaves("/replicate/dropdb?name=" + dbname)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Database dropped successfully"})
}

func createTable(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	dbname := r.URL.Query().Get("dbname")
	tableName := r.URL.Query().Get("table")

	// Handle request body for columns
	var tableRequest struct {
		Columns []Column `json:"columns"`
	}

	// Try to decode the request body into the struct
	if r.Body != nil {
		err := json.NewDecoder(r.Body).Decode(&tableRequest)
		if err != nil && err.Error() != "EOF" {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Check if required parameters are provided
	if dbname == "" || tableName == "" {
		http.Error(w, "Database name and table name are required", http.StatusBadRequest)
		return
	}

	// If no columns provided via JSON, check if there's a schema query parameter
	var columnsDef []string
	if len(tableRequest.Columns) == 0 {
		schema := r.URL.Query().Get("schema")
		if schema != "" {
			// Use the schema directly
			query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (%s)", dbname, tableName, schema)
			_, err := db.Exec(query)
			if err != nil {
				http.Error(w, "Failed to create table: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Either columns in request body or schema parameter is required", http.StatusBadRequest)
			return
		}
	} else {
		// Process columns from the request body
		for _, col := range tableRequest.Columns {
			columnsDef = append(columnsDef, fmt.Sprintf("`%s` %s", col.Name, col.DataType))
		}

		// Create table with primary key and columns
		createTableQuery := fmt.Sprintf(
			"CREATE TABLE IF NOT EXISTS `%s`.`%s` (Id INT AUTO_INCREMENT PRIMARY KEY, %s)",
			dbname, tableName, strings.Join(columnsDef, ", "),
		)

		_, err := db.Exec(createTableQuery)
		if err != nil {
			http.Error(w, "Failed to create table: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Replicate to slaves
	if len(tableRequest.Columns) > 0 {
		// Create a schema string for replication
		schema := fmt.Sprintf("Id INT AUTO_INCREMENT PRIMARY KEY, %s", strings.Join(columnsDef, ", "))
		go replicateToSlaves(fmt.Sprintf("/replicate/table?dbname=%s&table=%s&schema=%s",
			dbname, tableName, url.QueryEscape(schema)))
	} else {
		schema := r.URL.Query().Get("schema")
		go replicateToSlaves(fmt.Sprintf("/replicate/table?dbname=%s&table=%s&schema=%s",
			dbname, tableName, url.QueryEscape(schema)))
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Table created successfully"})
}

func insertRecord(w http.ResponseWriter, r *http.Request) {
	// Parse request data
	var req struct {
		DBName  string                 `json:"dbname"`
		Table   string                 `json:"table"`
		Values  string                 `json:"values"`
		Records map[string]interface{} `json:"records"` // For dynamic record insertion
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.Table == "" {
		http.Error(w, "Database name and table name are required", http.StatusBadRequest)
		return
	}

	// Handle different insertion methods
	if req.Values != "" {
		// Traditional way with raw values string
		query := fmt.Sprintf("INSERT INTO %s.%s VALUES (%s)", req.DBName, req.Table, req.Values)
		_, err := db.Exec(query)
		if err != nil {
			http.Error(w, "Failed to insert record: "+err.Error(), http.StatusInternalServerError)
			return
		}

		go replicateToSlavesJSON("/replicate/insert", req)
	} else if len(req.Records) > 0 {
		// Process dynamic record insertion
		columns := make([]string, 0, len(req.Records))
		placeholders := make([]string, 0, len(req.Records))
		values := make([]interface{}, 0, len(req.Records))

		for colName, value := range req.Records {
			columns = append(columns, "`"+colName+"`")
			placeholders = append(placeholders, "?")

			// Validate based on value type
			switch v := value.(type) {
			// في قسم معالجة القيم السلسلة (string values):
			case string:
				if len(v) > 0 {
					regex := regexp.MustCompile(`^[a-zA-Z0-9\s\-]{1,255}$`)
					if !regex.MatchString(v) {
						http.Error(w, fmt.Sprintf("Invalid value for %s. Must be valid characters (letters, numbers, spaces, or hyphens).", colName), http.StatusBadRequest)
						return
					}
				}
				decoded, err := base64.StdEncoding.DecodeString(v)
				if err != nil {
					http.Error(w, fmt.Sprintf("Invalid Base64 value for %s", colName), http.StatusBadRequest)
					return
				}
				values = append(values, string(decoded))
			case float64:
				// For numeric values
				if v <= 0 {
					http.Error(w, fmt.Sprintf("Invalid value for %s. Must be a positive number.", colName), http.StatusBadRequest)
					return
				}
			}

			values = append(values, value)
		}

		// Build and execute parameterized query
		query := fmt.Sprintf(
			"INSERT INTO `%s`.`%s` (%s) VALUES (%s)",
			req.DBName, req.Table, strings.Join(columns, ", "), strings.Join(placeholders, ", "),
		)

		_, err := db.Exec(query, values...)
		if err != nil {
			http.Error(w, "Failed to insert record: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create a request structure suitable for replication
		replicateReq := map[string]interface{}{
			"dbname":  req.DBName,
			"table":   req.Table,
			"records": req.Records,
		}

		go replicateToSlavesJSON("/replicate/insert", replicateReq)
	} else {
		http.Error(w, "Either 'values' or 'records' is required", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Record inserted successfully"})
}
func selectRecords(w http.ResponseWriter, r *http.Request) {
	dbname := r.URL.Query().Get("dbname")
	table := r.URL.Query().Get("table")

	if dbname == "" || table == "" {
		http.Error(w, "Both dbname and table parameters are required", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("SELECT * FROM %s.%s", dbname, table)
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Failed to query records: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		http.Error(w, "Failed to get columns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var results []map[string]interface{}
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			val := columnPointers[i].(*interface{})

			// تحويل []byte إلى string مباشرة
			if b, ok := (*val).([]byte); ok {
				row[col] = string(b) // هذه هي الخطوة الأهم!
			} else {
				row[col] = *val
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error during rows iteration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func updateRecord(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DBName string `json:"dbname"`
		Table  string `json:"table"`
		Set    string `json:"set"`
		Where  string `json:"where"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.Table == "" || req.Set == "" || req.Where == "" {
		http.Error(w, "All fields (dbname, table, set, where) are required", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("UPDATE %s.%s SET %s WHERE %s", req.DBName, req.Table, req.Set, req.Where)
	_, err := db.Exec(query)
	if err != nil {
		http.Error(w, "Failed to update record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	go replicateToSlavesJSON("/replicate/update", req)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Record updated successfully"})
}

func deleteRecord(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DBName string `json:"dbname"`
		Table  string `json:"table"`
		Where  string `json:"where"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.Table == "" || req.Where == "" {
		http.Error(w, "All fields (dbname, table, where) are required", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("DELETE FROM %s.%s WHERE %s", req.DBName, req.Table, req.Where)
	_, err := db.Exec(query)
	if err != nil {
		http.Error(w, "Failed to delete record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	go replicateToSlavesJSON("/replicate/delete", req)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Record deleted successfully"})
}

func replicateToSlaves(path string) {
	for _, addr := range slaveAddresses {
		go func(address string) {
			maxRetries := 3
			retryDelay := 2 * time.Second

			for i := 0; i < maxRetries; i++ {
				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Get(address + path)
				if err == nil {
					defer resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						log.Printf("Replication to %s succeeded", address)
						return
					}
				}

				if i < maxRetries-1 {
					log.Printf("Attempt %d failed for %s, retrying in %v...", i+1, address, retryDelay)
					time.Sleep(retryDelay)
					retryDelay *= 2
				}
			}
			log.Printf("Failed to replicate to %s after %d attempts", address, maxRetries)
		}(addr)
	}
}

func replicateToSlavesJSON(path string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal data for replication: %v", err)
		return
	}

	for _, addr := range slaveAddresses {
		go func(address string) {
			maxRetries := 3
			retryDelay := 2 * time.Second

			for i := 0; i < maxRetries; i++ {
				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Post(address+path, "application/json", strings.NewReader(string(jsonData)))
				if err == nil {
					defer resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						log.Printf("Replication to %s succeeded", address)
						return
					}
				}

				if i < maxRetries-1 {
					log.Printf("Attempt %d failed for %s, retrying in %v...", i+1, address, retryDelay)
					time.Sleep(retryDelay)
					retryDelay *= 2
				}
			}
			log.Printf("Failed to replicate to %s after %d attempts", address, maxRetries)
		}(addr)
	}
}

func startElection() {
	if electionInProgress {
		return
	}
	electionInProgress = true

	log.Println("Starting master election...")
	time.Sleep(time.Second * 2)

	if strings.HasSuffix(masterAddress, "8001") {
		promoteToMaster()
	}
}

func promoteToMaster() {
	isMaster = true
	masterAddress = "http://localhost:8001"
	log.Println("This node has been promoted to master")
}

func checkMasterHealth() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !isMaster {
			client := &http.Client{Timeout: 5 * time.Second}
			_, err := client.Get(masterAddress + "/ping")
			if err != nil {
				log.Printf("Master is down: %v", err)
				startElection()
			}
		}
	}
}
