package examples

import (
	"fmt"

	"github.com/lvank/termsuji/api"
)

// an example
func APIExample() {
	//Set a password from OGS > Settings > Account Settings
	if err := api.Authenticate("YourOGSUsername", "YourOGSPassword"); err != nil {
		//If a refresh token exists, the username and password are unused
		panic("Username and password incorrect")
	}
	fmt.Printf("Hello, %s!", api.AuthData.Username)
	fmt.Println("Active games:")
	games := api.GetGamesList()
	var game api.GameListData //store the last active game
	for i := range games.Games {
		game = games.Games[i]
		if game.GameOver() {
			//skip inactive games
			continue
		}
		fmt.Printf("%s\n==> %s\n", game.Name, game.Description())
	}
	fmt.Println("Single game:")
	gameDetails := api.GetGameData(game.ID)
	fmt.Printf("%#v", gameDetails)

	fmt.Println("Board state for that game:")
	boardState := api.GetGameState(game.ID)
	fmt.Printf("%#v", boardState)
}
