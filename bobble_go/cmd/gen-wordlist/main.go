package main

import (
	"archive/zip"
	"bufio"
	"compress/gzip"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	inpFile := flag.String("file", "", "File to load. One word per line.")
	dbPath := flag.String("db", "../words.db", "Path to SQLite database file.")
	flag.Parse()

	if *inpFile == "" {
		log.Fatalf("No input file specified. Use -file=<filename>")
	}

	words, err := loadWordList(*inpFile)
	if err != nil {
		log.Fatalf("Error loading word list: %v", err)
	}

	count, err := LoadWordsToSQLite(words, *dbPath, "words", "word")
	if err != nil {
		log.Fatalf("Error saving words to database: %v", err)
	}

	log.Printf("Successfully loaded %d words into the database", count)
}

func loadWordList(inputFile string) ([]string, error) {
	f, err := os.OpenFile(inputFile, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rdr io.ReadCloser = f
	finfo, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := finfo.Size()

	parts := strings.Split(inputFile, ".")
	extName := parts[len(parts)-1]

	log.Printf("Loading wordlist %s", inputFile) // Fixed variable name from fileName to inputFile

	switch extName {
	case "zip":
		var zipRdr *zip.Reader
		zipRdr, err = zip.NewReader(f, size)
		if err == nil {
			if len(zipRdr.File) == 1 && !zipRdr.File[0].FileInfo().IsDir() {
				rdr, err = zipRdr.File[0].Open()
			} else {
				err = errors.New("Zip file does not contain only one file")
			}
		}
	case "gz":
		fallthrough
	case "gzip":
		rdr, err = gzip.NewReader(f)
	case "txt":
		fallthrough
	default:
	}
	if err != nil {
		return nil, err
	}

	return ReadLines(rdr)
}

// ReadLines reads all lines from the provided reader and returns them as a slice of strings.
func ReadLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// loadWordsToSQLite loads a list of words into a SQLite database.
// Words are converted to uppercase and only those with length >= 3 are stored.
// The database will contain a single table with a single column that has a unique constraint.
func LoadWordsToSQLite(words []string, dbPath string, tableName string, columnName string) (int, error) {
	// Open or create the SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create the table if it doesn't exist
	createTableSQL := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				%s TEXT NOT NULL UNIQUE
			)
		`, tableName, columnName)

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return 0, fmt.Errorf("failed to create table: %w", err)
	}

	// Begin a transaction for faster inserts
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Prepare the insert statement
	insertSQL := fmt.Sprintf("INSERT OR IGNORE INTO %s (%s) VALUES (?)", tableName, columnName)
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// Insert words that meet the criteria
	insertedCount := 0
	for _, word := range words {
		// Convert to uppercase and check length
		upperWord := strings.ToUpper(word)
		if len(upperWord) >= 3 {
			_, err = stmt.Exec(upperWord)
			if err != nil {
				return insertedCount, fmt.Errorf("failed to insert word: %w", err)
			}
			insertedCount++
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return insertedCount, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return insertedCount, nil
}
