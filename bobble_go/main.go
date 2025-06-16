package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	bobble "bobble_go/game"

	_ "github.com/mattn/go-sqlite3"
)

// REST API Handlers

// HandleNewGame creates a new game
func HandleNewGame(w http.ResponseWriter, r *http.Request, db *sql.DB, defaultSize int) {
	// Get size from query parameter, use default if not provided
	sizeStr := r.URL.Query().Get("size")
	size := defaultSize

	if sizeStr != "" {
		var err error
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 4 || size > 7 {
			http.Error(w, "Invalid size parameter. Must be between 4 and 7", http.StatusBadRequest)
			return
		}
	}

	game := bobble.NewGame(size, db)

	// Respond with the board only, not all words
	response := struct {
		Size  int      `json:"size"`
		Board [][]rune `json:"board"`
	}{
		Size:  game.Size,
		Board: game.Board,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetAllWords returns all valid words on the board
func HandleGetAllWords(w http.ResponseWriter, r *http.Request, db *sql.DB, size int) {
	game := bobble.NewGame(size, db)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game.Words)
}

// HandleCheckWord checks if a specific word exists on the board
func HandleCheckWord(w http.ResponseWriter, r *http.Request, db *sql.DB, defaultSize int) {
	word := r.URL.Query().Get("word")
	if word == "" {
		http.Error(w, "Missing 'word' parameter", http.StatusBadRequest)
		return
	}

	// Get size from query parameter, use default if not provided
	sizeStr := r.URL.Query().Get("size")
	size := defaultSize

	if sizeStr != "" {
		var err error
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 4 || size > 7 {
			http.Error(w, "Invalid size parameter. Must be between 4 and 7", http.StatusBadRequest)
			return
		}
	}

	// Get board parameter if provided
	boardParam := r.URL.Query().Get("board")

	var game *bobble.Game
	if boardParam != "" {
		game = bobble.InitGame(size, boardParam, db)
	} else {
		game = bobble.NewGame(size, db)
	}

	result := game.HasWord(word)

	w.Header().Set("Content-Type", "application/json")
	if result != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"found": true,
			"word":  result.Text,
			"path":  result.Path,
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"found": false,
			"word":  word,
		})
	}
}

func main() {
	portFlag := flag.Int("port", 8080, "port to listen on")
	dbPathFlag := flag.String("db", "../words.db", "path to sqlite database")
	sizeFlag := flag.Int("size", 5, "number of rows/columns in the grid square")
	modeFlag := flag.String("mode", "server", "mode: server or cli")
	wordFlag := flag.String("word", "", "word to check (cli mode only)")
	flag.Parse()

	// Connect to the SQLite database
	db, err := sql.Open("sqlite3", *dbPathFlag)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if *modeFlag == "server" {
		// Add import for strconv at the top if not already there

		// Serve static files from the current directory
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.ServeFile(w, r, "assets/BobbleGame.html")
			} else {
				http.NotFound(w, r)
			}
		})

		// Server mode - start the REST API
		http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
			HandleNewGame(w, r, db, *sizeFlag)
		})

		http.HandleFunc("/words", func(w http.ResponseWriter, r *http.Request) {
			HandleGetAllWords(w, r, db, *sizeFlag)
		})

		http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
			HandleCheckWord(w, r, db, *sizeFlag)
		})

		addr := fmt.Sprintf(":%d", *portFlag)
		fmt.Printf("Bobble server starting on http://localhost%s\n", addr)
		fmt.Printf("Endpoints:\n")
		fmt.Printf("  - GET /new?size=[4-7] - Create a new game board\n")
		fmt.Printf("  - GET /words - Get all valid words on the board\n")
		fmt.Printf("  - GET /check?word=WORD - Check if a word exists on the board\n")
		log.Fatal(http.ListenAndServe(addr, nil))
	} else {
		// CLI mode
		game := bobble.NewGame(*sizeFlag, db)
		fmt.Println(game.Display())

		if *wordFlag != "" {
			// Check for a specific word
			result := game.HasWord(*wordFlag)
			if result != nil {
				fmt.Printf("Word '%s' found on the board!\n", *wordFlag)
				fmt.Printf("Path: %v\n", result.Path)
			} else {
				fmt.Printf("Word '%s' not found on the board.\n", *wordFlag)
			}
		} else {
			// Print all found words
			fmt.Printf("Found %d words on the board:\n", len(game.Words))
			for i, word := range game.Words {
				if i > 20 && len(game.Words) > 25 {
					fmt.Printf("... and %d more words\n", len(game.Words)-i)
					break
				}
				fmt.Printf("%s\n", word.Text)
			}
		}
	}
}
