package game

import (
	"database/sql"
	"log"
	"math/rand"
	"strings"
)

// NewGame creates a new game with a random board
func NewGame(size int, db *sql.DB) *Game {
	game := &Game{
		Size:  size,
		Board: make([][]rune, size),
		DB:    db,
	}

	numCells := size * size
	allDice := makeIntRange(len(BobbleDice))
	availableDice := makeIntRange(len(BobbleDice))
	if size == 3 {
		availableDice = []int{0, 1, 2, 3, 4, 5}
		allDice = []int{6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	}

	for len(availableDice) < numCells {
		if len(allDice) == 0 {
			allDice = makeIntRange(len(BobbleDice))
		}
		r := rand.Intn(len(allDice))
		d := allDice[r]
		allDice = append(allDice[:r], allDice[r+1:]...)
		availableDice = append(availableDice, d)
	}

	// Initialize the board with random letters
	for i := 0; i < size; i++ {
		game.Board[i] = make([]rune, size)
		for j := 0; j < size; j++ {
			// Get a random die and a random face
			r := rand.Intn(len(availableDice))
			d := availableDice[r]
			availableDice = append(availableDice[:r], availableDice[r+1:]...)
			die := BobbleDice[d]
			face := die[rand.Intn(len(die))]
			game.Board[i][j] = rune(face)
		}
	}

	// Find all valid words on the board
	// game.findAllWords()

	return game
}

// InitGame creates a new game with a given board
func InitGame(size int, board string, db *sql.DB) *Game {
	game := &Game{
		Size:  size,
		Board: make([][]rune, size),
		DB:    db,
	}

	for i := 0; i < size; i++ {
		game.Board[i] = make([]rune, size)
		for j := 0; j < size; j++ {
			game.Board[i][j] = rune(board[(i*size)+j])
		}
	}
	return game
}

func makeIntRange(i int) []int {
	list := make([]int, i)
	for j := 0; j < i; j++ {
		list[j] = j
	}
	return list
}

// findAllWords finds all valid words on the board
func (g *Game) findAllWords() {
	g.Words = []Word{}
	visited := make([][]bool, g.Size)
	for i := 0; i < g.Size; i++ {
		visited[i] = make([]bool, g.Size)
	}

	// Start DFS from each cell
	for i := 0; i < g.Size; i++ {
		for j := 0; j < g.Size; j++ {
			g.dfs(i, j, "", "", []Cell{}, visited)
			log.Printf("Finished %d:%d. Found %d words", i, j, len(g.Words))
		}
	}
}

// findAWord finds a valid word on the board
func (g *Game) findAWord(targetWord string) {
	g.Words = []Word{}
	visited := make([][]bool, g.Size)
	for i := 0; i < g.Size; i++ {
		visited[i] = make([]bool, g.Size)
	}

	// Start DFS from each cell
	for i := 0; i < g.Size; i++ {
		for j := 0; j < g.Size; j++ {
			g.dfs(i, j, "", targetWord, []Cell{}, visited)
			log.Printf("Finished %d:%d. Found %d words", i, j, len(g.Words))
		}
	}
}

// isValidWord checks if a word exists in the database
func (g *Game) isValidWord(word string) (bool, bool) {
	var ret string
	//var count int
	//err := g.DB.QueryRow("SELECT COUNT(*) FROM words WHERE word = ?", strings.ToLower(word)).Scan(&count)
	err := g.DB.QueryRow("SELECT word FROM words WHERE word like ? ORDER BY LENGTH(word) ASC", strings.ToLower(word+"%")).Scan(&ret)
	if err != nil {
		if err.Error() != "sql: no rows in result set" {
			log.Printf("Error checking word validity: %v", err)
			return false, false
		}
	}
	return word == ret, strings.Index(ret, word) == 0
}

// dfs performs depth-first search to find words on the board
func (g *Game) dfs(row, col int, currentWord, targetWord string, path []Cell, visited [][]bool) {
	// Check boundaries and if already visited
	if row < 0 || row >= g.Size || col < 0 || col >= g.Size || visited[row][col] {
		return
	}

	// Add current letter to word
	currentLetter := g.Board[row][col]
	currentWord = currentWord + string(currentLetter)
	currentPath := append(path, Cell{Row: row, Col: col})

	if len(targetWord) > 0 && strings.Index(targetWord, currentWord) != 0 {
		return
	}

	// Mark as visited for this path
	visited[row][col] = true

	// Default to valid prefix until length great enough to check word
	var validWord, validPrefix = false, true

	// If the word is valid and has at least 3 letters, add it to results
	if len(currentWord) >= 3 {
		if len(targetWord) > 0 {
			validPrefix = true // we checked this above
			validWord = currentWord == targetWord
			if validWord {
				validWord, validPrefix = g.isValidWord(currentWord)
			}
		} else {
			validWord, validPrefix = g.isValidWord(currentWord)
		}
		if validWord {
			path = make([]Cell, 0, len(currentPath))
			path = append(path, currentPath...)
			g.Words = append(g.Words, Word{
				Text: currentWord,
				Path: path,
			})
		}
	}

	if validPrefix {
		// Explore all 8 adjacent cells
		for dr := -1; dr <= 1; dr++ {
			for dc := -1; dc <= 1; dc++ {
				if dr != 0 || dc != 0 {
					g.dfs(row+dr, col+dc, currentWord, targetWord, currentPath, visited)
				}
			}
		}
	}

	// Backtrack
	visited[row][col] = false
}

// HasWord checks if a specific word exists in the found words
func (g *Game) HasWord(word string) *Word {
	if len(g.Words) == 0 {
		g.findAWord(word)
	}
	for _, w := range g.Words {
		if strings.ToUpper(w.Text) == strings.ToUpper(word) {
			return &w
		}
	}
	return nil
}

// Display board in text format
func (g *Game) Display() string {
	var sb strings.Builder
	sb.WriteString("Bobble Board:\n")
	for i := 0; i < g.Size; i++ {
		for j := 0; j < g.Size; j++ {
			sb.WriteRune(g.Board[i][j])
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
