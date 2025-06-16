## Initial Prompt
`Implement an API that has both a REST interface and is callable via commandline application. The API implements a game like Boggle, that I call bobble, and has a configurable sized square grid of random letters from the pool of letters available to Boggle. It also has a call that returns all possible word combinations that exist in that square. Each returned item should have the word and the ordered list of cells that contain the word. Another call will take a word as a parameter and return if it is in the list of words in the square. The list of words come from a sqlite db as defined in cmd/gen-wordlist/main.go`

### Notes
The AI (Claude 3.7 Sonnet) failed to calculate the dice properly, choosing to place the same die into multiple places by leaving the die in the selection pool.

It also made a poor choice to not abandon the word search early if the current letters are not a valid prefix to a valid word.

It made several other smaller errors also.

## Second Prompt
`Add to "/new" to take a size parameter. Use that size to create a new game. Add a html file that asks for size in a select control that has 4 through 7 as options and calls "/new" and shows the board as a grid. Add a search that calls "/check" and displays if it is an allowed word or not. Title the page "BOBBLE".`

### Notes
AI chose to not implement reloading the game from a specified board so I added that and modified the JS to send it.
Along with this, I changed the find-word functionality to optimize looking through the board for a particular word so it did not have to spend minutes finding all possible words just to validate one.

## Third Prompt
`Add a timer on the BobbleGame.html that counts down seconds from 1 minute. It should start after the New Game board is shown. Make sure the page background is white. When it hits 0, change the page background to light grey and show an alert that time has elapsed.`

### Notes
