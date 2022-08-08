package examples

import (
	"fmt"

	"github.com/lvank/termsuji/api"
)

// an example
func APIExample() {
	//You can only authenticate with a refresh token if you have one stored from an earlier password authentication.
	if err := api.AuthenticateRefreshToken("YourStoredRefreshToken"); err != nil {
		//Set a password from OGS > Settings > Account Settings
		if err = api.AuthenticatePassword("YourOGSUsername", "YourOGSPassword"); err != nil {
			panic("Username and password incorrect")
		}
	}
	fmt.Printf("Hello, %s!", api.AuthData.Player.Username)
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
