package api

import (
	"fmt"

	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

//A socket client wrapper for communicating with the realtime API
type RealtimeClient struct {
	c      *gosocketio.Client
	GameID int64
}

var (
	wsUrl = "wss://online-go.com/socket.io/?EIO=3&transport=websocket"
)

type EmitAuth struct {
	Auth     string `json:"auth"`
	PlayerID int64  `json:"player_id"`
	Username string `json:"username"`
}

type EmitGameConnect struct {
	GameID   int64 `json:"game_id"`
	PlayerID int64 `json:"player_id"`
	Chat     bool  `json:"chat"`
}

type EmitMove struct {
	GameID   int64  `json:"game_id"`
	PlayerID int64  `json:"player_id"`
	Move     string `json:"move"`
}

//Create a new RealtimeClient which opens a socket connection for a specific game ID.
//The RealtimeClient is currently made to connect to one game at a time.
//An optional function f may be provided that will get called whenever the "game/<id>/gamedata"
//event is received, which happens directly after connecting and during the stone removal/finished phases.
//You are responsible for calling Disconnect() when the RealtimeClient is no longer required.
func Connect(gameID int64, f func(interface{})) (*RealtimeClient, error) {
	var r *RealtimeClient = &RealtimeClient{GameID: gameID}
	c, err := gosocketio.Dial(wsUrl, transport.GetDefaultWebsocketTransport())
	if err != nil {
		return nil, err
	}
	if f != nil {
		aFunc := func(i interface{}, response map[string]interface{}) {
			f(response)
		}
		c.On(fmt.Sprintf("game/%d/gamedata", gameID), aFunc)
	}
	c.Emit("game/connect", &EmitGameConnect{
		GameID:   r.GameID,
		PlayerID: AuthData.Player.ID,
		Chat:     false,
	})
	r.c = c
	return r, nil
}

//OnMoveResult is used as a return for the OnMove callback event.
type OnMoveResult struct {
	GameID     int64    `json:"game_id"`
	Move       BoardPos `json:"move"`
	MoveNumber int      `json:"move_number"`
}

//OnMove registers a callback for whenever a move is played in the connected game.
func (r *RealtimeClient) OnMove(f func(OnMoveResult)) {
	aFunc := func(i interface{}, response OnMoveResult) {
		f(response)
	}
	r.c.On(fmt.Sprintf("game/%d/move", r.GameID), aFunc)
}

type OnClockResult struct {
	CurrentPlayerID int64 `json:"current_player"`
	BlackPlayerID   int64 `json:"black_player_id"`
	WhitePlayerID   int64 `json:"white_player_id"`
	LastMove        int64 `json:"last_move"` //ms since epoch
	Now             int64 `json:"now"`       //ms since epoch
}

func (r *RealtimeClient) OnClock(f func(OnClockResult)) {
	aFunc := func(i interface{}, response OnClockResult) {
		f(response)
	}
	r.c.On(fmt.Sprintf("game/%d/clock", r.GameID), aFunc)
}

//Authenticate gets a token from the REST API which is submitted through the Realtime API websocket.
//This is required before calling authenticated functions, like RealtimeClient.Move.
//This function requires being authenticated through api.Authenticate first.
func (r *RealtimeClient) Authenticate() {
	auth := GetOGSConfig().ChatAuth
	r.c.Emit("authenticate", &EmitAuth{
		Auth:     auth,
		Username: AuthData.Player.Username,
		PlayerID: AuthData.Player.ID,
	})
}

func (r *RealtimeClient) Move(x, y int) {
	r.c.Emit("game/move", &EmitMove{
		GameID:   r.GameID,
		PlayerID: AuthData.Player.ID,
		Move:     PosSGF(BoardPos{X: x, Y: y}),
	})
}

//Disconnect closes the underlying websocket.
func (r *RealtimeClient) Disconnect() {
	r.c.Close()
}
