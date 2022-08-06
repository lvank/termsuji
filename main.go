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
	api.Authenticate("", "") //attempt refresh token auth
	cfg := config.InitConfig()
	cfg.Save() // TODO settings screen or something
	app := tview.NewApplication()
	rootPage := tview.NewPages()
	rootPage.SetBorder(true).SetTitle("termsuji")
	list := tview.NewList()

	gameBoard := ui.NewGoBoard(app, cfg)
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
		AddInputField("Username", api.AuthData.Username, 32, nil, nil).
		AddPasswordField("Password", "", 32, '*', nil).
		AddButton("Submit", func() {
			err := api.Authenticate(
				loginForm.GetFormItem(0).(*tview.InputField).GetText(),
				loginForm.GetFormItem(1).(*tview.InputField).GetText(),
			)
			if err != nil {
				loginFrame.Clear().AddText(err.Error(), true, tview.AlignLeft, tcell.PaletteColor(1))
				return
			}
			refreshGames(rootPage, list, gameBoard)
			rootPage.SwitchToPage("browser")
		})
	loginFrame.
		SetBorders(0, 0, 0, 0, 1, 0).
		AddText("Log in to OGS", true, tview.AlignLeft, tcell.PaletteColor(3))

	rootPage.AddPage("login", loginFrame, true, true)
	rootPage.AddPage("browser", list, true, false)
	rootPage.AddPage("gameview", gameBoard.Box, true, false)

	if api.AuthData.Authenticated {
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

func mockBoard() *api.BoardData {
	return &api.BoardData{
		Width:         19,
		Height:        19,
		InitialPlayer: "black",
		Moves: []api.BoardPos{
			{X: 0, Y: 0},
			{X: 0, Y: 18},
			{X: 18, Y: 0},
			{X: 18, Y: 18},
		},
	}
}
