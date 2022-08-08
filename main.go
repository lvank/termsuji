//termsuji is a application to play online-go.com in a terminal.
//It is not complete and can only read board state and play moves, and is
//intended more as a reference than a full-fledged application.
package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/lvank/termsuji/api"
	"github.com/lvank/termsuji/config"
	"github.com/lvank/termsuji/ui"
	"github.com/rivo/tview"
)

func main() {
	auth := config.InitAuthData()
	if auth.Tokens.Refresh != "" {
		api.AuthenticateRefreshToken(auth.Tokens.Refresh)
	}
	cfg := config.InitConfig()
	cfg.Save() // TODO settings screen or something
	app := tview.NewApplication()
	rootPage := tview.NewPages()
	rootPage.SetBorder(true).SetTitle("termsuji")
	list := tview.NewList()

	gameFrame := tview.NewFlex()
	gameHint := tview.NewTextView()
	gameHint.SetBorder(true)
	gameBoard := ui.NewGoBoard(app, cfg, gameHint)
	gameFrame.
		AddItem(gameBoard.Box, 19*2+2, 1, true).
		AddItem(gameHint, 0, 2, false)
	gameBoard.Box.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			if gameBoard.SelectedTile() != nil {
				gameBoard.ResetSelection()
			} else {
				gameBoard.Close()
				rootPage.SwitchToPage("browser")
			}
			return nil
		}
		switch event.Key() {
		case tcell.KeyUp:
			gameBoard.MoveSelection(0, -1)
		case tcell.KeyDown:
			gameBoard.MoveSelection(0, 1)
		case tcell.KeyLeft:
			gameBoard.MoveSelection(-1, 0)
		case tcell.KeyRight:
			gameBoard.MoveSelection(1, 0)
		case tcell.KeyEnter:
			selTile := gameBoard.SelectedTile()
			if selTile == nil {
				return nil
			}
			gameBoard.PlayMove(selTile.X, selTile.Y)
		case tcell.KeyRune:
			if event.Rune() == 'p' {
				gameBoard.PlayMove(-1, -1)
			}
		}
		return event
	})
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			app.Stop()
			return nil
		}
		return event
	})

	loginForm := tview.NewForm()
	loginFrame := tview.NewFrame(loginForm)
	loginForm.
		AddInputField("Username", auth.Username, 32, nil, nil). //if we have a cached username, prefill it
		AddPasswordField("Password", "", 32, '*', nil).
		AddButton("Submit", func() {
			err := api.AuthenticatePassword(
				loginForm.GetFormItem(0).(*tview.InputField).GetText(),
				loginForm.GetFormItem(1).(*tview.InputField).GetText(),
			)
			if err != nil {
				loginFrame.Clear().AddText(err.Error(), true, tview.AlignLeft, tcell.PaletteColor(1))
				return
			}
			storeAuthData(auth)
			refreshGames(rootPage, list, gameBoard)
			rootPage.SwitchToPage("browser")
		})
	loginFrame.
		SetBorders(0, 0, 0, 0, 1, 0).
		AddText("Log in to OGS", true, tview.AlignLeft, tcell.PaletteColor(3))

	rootPage.AddPage("login", loginFrame, true, true)
	rootPage.AddPage("browser", list, true, false)
	rootPage.AddPage("gameview", gameFrame, true, false)

	if api.AuthData.Authenticated {
		storeAuthData(auth)
		refreshGames(rootPage, list, gameBoard)
		rootPage.SwitchToPage("browser")
	} else {
		rootPage.SwitchToPage("login")
	}

	if err := app.SetRoot(rootPage, true).Run(); err != nil {
		panic(err)
	}
}

func refreshGames(root *tview.Pages, list *tview.List, gui *ui.GoBoardUI) {
	list.Clear()
	gamelist := api.GetGamesList()
	i := 0
	for _, game := range gamelist.Games {
		if game.GameOver() {
			continue
		}
		gameID := game.ID
		list.AddItem(game.Name, game.Description(), rune('a'+i), func() {
			gui.Connect(gameID)
			root.SwitchToPage("gameview")
		})
		i++
	}
}

//Stores authentication data from api package after successful authentication.
func storeAuthData(a *config.AuthData) {
	a.Username = api.AuthData.Player.Username
	a.UserID = api.AuthData.Player.ID
	a.Tokens.Refresh = api.AuthData.Oauth.RefreshToken
	a.Save()
}
