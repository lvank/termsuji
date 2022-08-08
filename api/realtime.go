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
//An optional function f may be provided that will get called whenever the "game/<id>/gamedata"
//event is received, which happens directly after connecting and during the stone removal/finished phases.
//You are responsible for calling Disconnect() when the RealtimeClient is no longer required.
func Connect(gameID int64, f interface{}) (*RealtimeClient, error) {
	var r *RealtimeClient = &RealtimeClient{GameID: gameID}
	c, err := gosocketio.Dial(wsUrl, transport.GetDefaultWebsocketTransport())
	if err != nil {
		return nil, err
	}
	if f != nil {
		r.c.On(fmt.Sprintf("game/%d/gamedata", gameID), f)
	}
	c.Emit("game/connect", &EmitGameConnect{
		GameID:   r.GameID,
		PlayerID: AuthData.Player.ID,
		Chat:     false,
	})
	r.c = c
	return r, nil
}

//OnMove registers a callback for whenever a move is played in the connected game.
func (r *RealtimeClient) OnMove(f interface{}) {
	r.c.On(fmt.Sprintf("game/%d/move", r.GameID), f)
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
