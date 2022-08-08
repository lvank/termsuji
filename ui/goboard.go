//Package ui specifies custom controls for tview to assist in playing Go in the terminal.
package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lvank/termsuji/api"
	"github.com/lvank/termsuji/config"
	"github.com/rivo/tview"
)

type GoBoardUI struct {
	Box          *tview.Box
	BoardState   *api.BoardState
	hint         *tview.TextView
	selX         int
	selY         int
	lastTurnPass bool
	app          *tview.Application
	rc           *api.RealtimeClient
	styles       []tcell.Color
}

func (g *GoBoardUI) SelectedTile() *api.BoardPos {
	if g.selX == -1 && g.selY == -1 {
		return nil
	}
	return &api.BoardPos{X: g.selX, Y: g.selY}
}

func (g *GoBoardUI) MoveSelection(h, v int) {
	prevTile := g.SelectedTile()
	if prevTile == nil {
		g.selX = g.BoardState.LastMove.X
		g.selY = g.BoardState.LastMove.Y
		if g.SelectedTile() == nil {
			//no previous move made, use board center
			g.selX = int(g.BoardState.Width() / 2)
			g.selY = int(g.BoardState.Height() / 2)
		}
		return
	}
	if g.selX+h < 0 || g.selX+h >= g.BoardState.Width() {
		return
	}
	if g.selY+v < 0 || g.selY+v >= g.BoardState.Width() {
		return
	}
	g.selX += h
	g.selY += v
}

func (g *GoBoardUI) ResetSelection() {
	g.selX = -1
	g.selY = -1
}

func NewGoBoard(app *tview.Application, c *config.Config, hint *tview.TextView) *GoBoardUI {
	goBoard := &GoBoardUI{
		Box:        tview.NewBox(),
		BoardState: &api.BoardState{},
		hint:       hint,
		app:        app,
		selX:       -1,
		selY:       -1,
		styles: []tcell.Color{
			tcell.PaletteColor(c.BoardColor),
			tcell.PaletteColor(c.BlackColor),
			tcell.PaletteColor(c.WhiteColor),
			tcell.PaletteColor(c.BoardColorAlt),
			tcell.PaletteColor(c.BlackColorAlt),
			tcell.PaletteColor(c.WhiteColorAlt),
			tcell.PaletteColor(c.CursorColor),
		},
	}
	goBoard.Box.SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
		if goBoard.BoardState == nil {
			return x, y, 1, 1
		}
		//To approximate squares, double the width of characters
		boardW, boardH := goBoard.BoardState.Width()*2, goBoard.BoardState.Height()

		for boardY := 0; boardY < goBoard.BoardState.Height(); boardY++ {
			for boardX := 0; boardX < goBoard.BoardState.Width(); boardX++ {
				i := goBoard.BoardState.Board[boardY][boardX]
				iInv := 0
				//Cursor color is inverted stone color, or cursor color when not on a stone.
				var fgColor tcell.Color
				if i == 1 {
					iInv = 2
				} else if i == 2 {
					iInv = 1
				}
				if (boardX%2 + boardY%2) == 1 {
					i += 3
					if iInv > 0 {
						iInv += 3
					}
				}
				if iInv > 0 {
					//there's a stone
					fgColor = goBoard.styles[iInv]
				} else {
					//no stone, use cursor color
					fgColor = goBoard.styles[6]
				}
				drawRune := ' '
				if boardX == goBoard.BoardState.LastMove.X && boardY == goBoard.BoardState.LastMove.Y {
					drawRune = '/'
				}
				if boardX == goBoard.selX && boardY == goBoard.selY {
					drawRune = 'X'
				}

				drawCell(screen, tcell.StyleDefault.Background(goBoard.styles[i]).Foreground(fgColor), drawRune, boardX, boardY, x, y)
			}
		}

		return x, y, boardW, boardH
	})
	return goBoard
}

func (g *GoBoardUI) Connect(gameID int64) {
	realtimeClient, err := api.Connect(gameID, nil)
	g.rc = realtimeClient
	if err != nil {
		panic(err)
	}
	g.rc.Authenticate()
	g.rc.OnMove(func(m api.OnMoveResult) {
		//If X/Y are -1, the last turn was a pass.
		g.lastTurnPass = (m.Move.X == -1 && m.Move.Y == -1)
		g.refreshBoard()
		g.app.QueueUpdateDraw(func() {})
	})
	g.rc.OnClock(func(c api.OnClockResult) {
		if c.CurrentPlayerID == api.AuthData.Player.ID {
			g.setHint("It is your turn.")
		} else {
			g.setHint("It is your opponent's turn.")
		}
	})
	g.refreshBoard()
}

func (g *GoBoardUI) PlayMove(x, y int) {
	g.rc.Move(x, y)
}

func (g *GoBoardUI) Close() {
	if g.rc == nil {
		return
	}
	g.rc.Disconnect()
}

func (g *GoBoardUI) refreshBoard() {
	g.BoardState = api.GetGameState(g.rc.GameID)
	if g.BoardState.PlayerToMove == api.AuthData.Player.ID {
		g.setHint("It is your turn.")
	} else {
		g.setHint("It is your opponent's turn.")
	}
}

func (g *GoBoardUI) setHint(hint string) {
	var passHint string
	if g.lastTurnPass {
		passHint = "The previous turn was passed.\n\n"
	}
	g.hint.SetText(fmt.Sprintf("%s%s\n\narrow keys: move cursor, Return: play move\nBackspace: pass turn, q: quit", passHint, hint))
}

//Helper function to draw a single cell, which occupies two characters on screen
func drawCell(s tcell.Screen, c tcell.Style, r rune, x, y, l, t int) {
	for i := 0; i < 2; i++ {
		s.SetContent(l+x*2+i, t+y, r, nil, c)
	}
}
