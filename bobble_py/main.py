from fastapi import FastAPI, Depends
from fastapi.params import Query
from starlette.responses import HTMLResponse
from sqlalchemy.orm import Session

from bobble import bobble
from database import get_db
app = FastAPI()


@app.get("/",  response_class=HTMLResponse)
async def read_index():
    with open("../web/BobbleGame.html", "r") as f:
        return HTMLResponse(f.read())


@app.get("/new", response_model=bobble.Game)
async def new_game(db: Session = Depends(get_db), size: int = Query(...)):
    if size == 0:
        size = 5
    if size < 4 or size > 7:
        raise ValueError("Size must be between 4 and 7")

    game = bobble.new(size)
    return game

@app.get("/check", response_model=dict)
async def check(db: Session = Depends(get_db), word: str = Query(...), size: int = Query(...), board: str = Query(...)):
    game = bobble.init(size, board)
    result = game.has_word(word, db)
    if result is not None :
        return dict(
            found=True,
            word=result.word,
            path=result.path,
        )
    return dict(
        found=False,
        word=word,
    )
if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)