package game

import "database/sql"

// Cell represents a position on the game board
type Cell struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// Word represents a word found on the board and its path
type Word struct {
	Text string `json:"word"`
	Path []Cell `json:"path"`
}

// Game represents the Bobble game
type Game struct {
	Size  int      `json:"size"`
	Board [][]rune `json:"board"`
	Words []Word   `json:"words,omitempty"`
	DB    *sql.DB
}

// BobbleDice represents the distribution of letters on each die
var BobbleDice = []string{
	"AAEEGN", "ABBJOO", "ACHOPS", "AFFKPS",
	"AOOTTW", "CIMOTU", "DEILRX", "DELRVY",
	"DISTTY", "EEGHNW", "EEINSU", "EHRTVW",
	"EIOSST", "ELRTTY", "HIMNQU", "HLNNRZ",
}
