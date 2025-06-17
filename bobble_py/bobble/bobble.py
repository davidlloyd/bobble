import random

from pydantic import BaseModel
from typing import List, Optional

from sqlalchemy import Select, FromClause, select, String
from sqlalchemy.orm import Session, DeclarativeBase, MappedColumn, Mapped


class Base(DeclarativeBase):
    pass

class Word(Base):
    __tablename__ = "words"
    word: Mapped[str] = MappedColumn(String(250), primary_key=True)

"""BobbleDice represents the distribution of letters on each die"""
BobbleDice = [
    "AAEEGN", "ABBJOO", "ACHOPS", "AFFKPS",
    "AOOTTW", "CIMOTU", "DEILRX", "DELRVY",
    "DISTTY", "EEGHNW", "EEINSU", "EHRTVW",
    "EIOSST", "ELRTTY", "HIMNQU", "HLNNRZ",
    ]

class Cell(BaseModel):
    row: int
    col: int
    def __init__(self, **data):
        super().__init__(**data)

class WordPath(BaseModel):
    word: str
    path: List[Cell]
    def __init__(self, **data):
        super().__init__(**data)

class Game(BaseModel):
    size: int = 5
    board: List[List[int]] = []
    words: List[WordPath] = []

    def __init__(self, **data):
        super().__init__(**data)
        if not self.words:
            self.words = []

    def has_word(self, word: str, db) -> Optional[WordPath]:
        self.words = []
        visited =[]
        for i in range(self.size) :
            visited.append([])
            for j in range(self.size) :
                visited[i].append(False)

        # Start DFS from each cell
        for i in range(self.size):
            for j in range(self.size):
                self._dfs(i, j, "", word, [], visited, db)

        return self.words[0] if len(self.words) > 0 else None

    def _dfs(self, row, col, current_word, target_word, path, visited, db: Session):
        # Check boundaries and if already visited
        if row < 0 or row >= self.size or col < 0 or col >= self.size or visited[row][col] :
            return

        # Add current letter to word
        current_letter = self.board[row][col]
        current_word = current_word + chr(current_letter)
        current_path = path.copy()
        current_path.append(Cell(row=row, col=col))

        if len(target_word) > 0 and not target_word.startswith(current_word):
            return

        # Mark as visited for this path
        visited[row][col] = True

        # Default to valid prefix until length great enough to check word
        valid_word, valid_prefix = False, True

        # If the word is valid and has at least 3 letters, add it to results
        if len(current_word) >= 3 :
            if len(target_word) > 0 :
                valid_prefix = True # we checked this above
                valid_word = current_word == target_word
                if valid_word :
                    valid_word, valid_prefix = self.is_valid_word(current_word, db)
            else :
                valid_word, valid_prefix = self.is_valid_word(current_word, db)

            if valid_word :
                self.words.append(WordPath(word=current_word, path=current_path))

        if valid_prefix :
            # Explore all 8 adjacent cells
            for dr in range(-1, 2):
                for dc in range(-1, 2):
                    if dr != 0 or dc != 0 :
                        self._dfs(row+dr, col+dc, current_word, target_word, current_path, visited, db)

        # Backtrack
        visited[row][col] = False

    def is_valid_word(self, aword: str, db: Session):
        # Query for exact word match
        stmt = select(Word).where(Word.word.__eq__(aword))
        word_match = db.execute(stmt).fetchone()

        match = False
        prefix = False

        if word_match and hasattr(word_match, 'Word'):
            row = word_match[0]
            match = row.word == aword
            prefix = True
        else:
            # Query for prefix match
            prefix_stmt = select(Word).where(Word.word.startswith(aword))
            prefix_match = db.execute(prefix_stmt).fetchone()
            prefix = prefix_match is not None and hasattr(prefix_match, 'Word') and prefix_match[0].word!=""

        return bool(match), bool(prefix)


def new(size: int):
    num_cells = size * size
    all_dice = make_int_range(len(BobbleDice))
    available_dice = make_int_range(len(BobbleDice))
    if size == 3:
        available_dice = [0, 1, 2, 3, 4, 5]
        all_dice = [6, 7, 8, 9, 10, 11, 12, 13, 14, 15]

    while len(available_dice) < num_cells:
        if len(all_dice) == 0:
            all_dice = make_int_range(len(BobbleDice))

        r = random.randint(0, len(all_dice) -1)
        d = all_dice[r]
        del all_dice[r]
        available_dice.append(d)

    board = []
    for row in range(size):
        board.append([])
        for col in range(size):
            r = random.randint(0, len(available_dice)-1)
            d = available_dice[r]
            del available_dice[r]
            die = BobbleDice[d]
            face = die[random.randint(0, len(die)-1)]
            board[row].append(ord(face))

    return Game(size=size, board=board)

def init(size: int, board_string: str):
    board_string = board_string.upper()
    board = []
    for row in range(size):
        board.append([])
        for col in range(size):
            board[row].append( ord(board_string[(row * size) + col]))
    return Game(size=size, board=board)

def make_int_range(n: int):
    r = []
    for i in range(n):
        r.append(i)
    return r