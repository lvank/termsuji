//termsuji is a application to play online-go.com in a terminal.
//It is not complete and can only read board state and play moves, and is
//intended more as a reference than a full-fledged application.
package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lvank/termsuji/api"
	"github.com/lvank/termsuji/config"
	"github.com/lvank/termsuji/ui"
	"github.com/rivo/tview"
)

var lastRefresh time.Time = time.Now()
var app *tview.Application
var rootPage *tview.Pages
var gameListFrame *tview.Frame
var gameList *tview.List
var frameHint *tview.Frame
var gameBoard *ui.GoBoardUI
var setLoading func(bool)

func main() {
	auth := config.InitAuthData()
	if auth.Tokens.Refresh != "" {
		api.AuthenticateRefreshToken(auth.Tokens.Refresh)
	}
	cfg, err := config.InitConfig()
	if err != nil {
		panic(err)
	}
	cfg.Save() // TODO settings screen or something
	app = tview.NewApplication()
	rootPage = tview.NewPages()
	rootPage.SetBorder(true).SetTitle("termsuji")
	gameList = tview.NewList()
	gameListFrame = tview.NewFrame(gameList)
	gameListFrame.SetBorders(0, 0, 0, 0, 0, 0)

	loadingModal := tview.NewModal()
	loadingModal.SetText("Loading...")
	setLoading = func(loading bool) {
		f := rootPage.ShowPage
		if !loading {
			f = rootPage.HidePage
		}
		f("loading")
	}
	gameFrame := tview.NewFlex()
	gameHint := tview.NewTextView()
	gameHint.SetBorder(true)
	gameBoard = ui.NewGoBoard(app, cfg, gameHint)
	gameFrame.
		AddItem(gameBoard.Box, 20*2+3, 1, true).
		AddItem(gameHint, 0, 2, false)
	gameBoard.Box.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			if gameBoard.SelectedTile() != nil {
				gameBoard.ResetSelection()
			} else {
				gameBoard.Close()
				async(func() {
					refreshGames()
					rootPage.SwitchToPage("browser")
				})
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
			switch event.Rune() {
			case 'p':
				gameBoard.PlayMove(-1, -1)
			case 't':
				rootPage.ShowPage("themes")
			}
		}
		return event
	})
	gameList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			case 'r':
				async(func() {
					refreshGames()
				})
				return nil
			case 't':
				rootPage.ShowPage("themes")
				return nil
			}
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
			refreshGames()
			rootPage.SwitchToPage("browser")
		})
	loginFrame.
		SetBorders(0, 0, 0, 0, 1, 0).
		AddText("Log in to OGS", true, tview.AlignLeft, tcell.PaletteColor(3))

	themeList := tview.NewList()
	themeList.SetTitle("Choose a theme")
	selectTheme := func(i int, main, secondary string, shortcut rune) {
		var theme config.Theme
		switch main {
		case "default":
			theme = config.DefaultTheme
		case "vaporwave":
			theme = config.VaporwaveTheme
		case "unicode":
			theme = config.UnicodeTheme
		case "catdog":
			theme = config.CatdogTheme
		case "hongoku":
			theme = config.HongokuTheme
		default: //No action
		}

		if main != "quit" {
			cfg.Theme = theme
			cfg.Save()
			gameBoard.SetConfig(cfg)
		}
		rootPage.HidePage("themes")
	}
	themeList.
		AddItem("default", "The default board theme", '1', nil).
		AddItem("vaporwave", "Magenta/cyan board theme", '2', nil).
		AddItem("unicode", "Display board with Unicode emoji symbols", '3', nil).
		AddItem("catdog", "Board becomes zoo", '4', nil).
		AddItem("hongoku", "Board becomes unreadable", '5', nil).
		AddItem("quit", "Keep your current settings", 'q', nil).
		SetSelectedFunc(selectTheme)

	rootPage.AddPage("login", loginFrame, true, true)
	rootPage.AddPage("browser", gameListFrame, true, false)
	rootPage.AddPage("gameview", gameFrame, true, false)
	rootPage.AddPage("themes", themeList, true, false)
	rootPage.AddPage("loading", loadingModal, false, false)

	if api.AuthData.Authenticated {
		storeAuthData(auth)
		refreshGames()
		rootPage.SwitchToPage("browser")
	} else {
		rootPage.SwitchToPage("login")
	}

	if err := app.SetRoot(rootPage, true).Run(); err != nil {
		panic(err)
	}
}

func refreshGames() {
	gameList.Clear()
	async(func() {
		lastRefresh = time.Now()
		gamesArray := api.GetGamesList()
		i := 0
		for _, game := range gamesArray.Games {
			if game.GameOver() {
				continue
			}
			gameID := game.ID
			gameList.AddItem(game.Name, game.Description(), rune('1'+i), func() {
				async(func() {
					gameBoard.Connect(gameID)
					rootPage.SwitchToPage("gameview")
				})
			})
			i++
		}
		gameListHint := fmt.Sprintf("r: refresh, t: themes, q: quit")
		gameListFrame.Clear().AddText(gameListHint, false, tview.AlignLeft, tcell.ColorDefault)
	})
}

//Helper function to show a loading screen while blocking functions are being called.
func async(f func()) {
	go func() {
		setLoading(true)
		app.Draw()
		f()
		setLoading(false)
		app.Draw()
	}()
}

//Stores authentication data from api package after successful authentication.
func storeAuthData(a *config.AuthData) {
	a.Username = api.AuthData.Player.Username
	a.UserID = api.AuthData.Player.ID
	a.Tokens.Refresh = api.AuthData.Oauth.RefreshToken
	a.Save()
}
