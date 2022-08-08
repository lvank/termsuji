//Package api contains methods to interact with the online-go.com REST API and Realtime API.
//For methods that require authentication, populate api.OauthClientID and call AuthenticatePassword
//or AuthenticateRefreshToken first before using the rest of the API. If successful, AuthData will contain
//relevant information about the user and tokens.
package api

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var (
	//go:embed client_id.txt
	oauthClientIDRaw string //may contain whitespace or other characters; use the exported one instead
	OauthClientID    = strings.TrimSpace(oauthClientIDRaw)
	BaseURL          = "https://online-go.com"
	AuthData         UserInfo

	//errors
	InvalidRefreshToken = errors.New("Invalid refresh token")

	apiURL                  = fmt.Sprintf("%s/api/v1/", BaseURL)
	termApiURL              = fmt.Sprintf("%s/termination-api/", BaseURL)
	oauthURL                = fmt.Sprintf("%s/oauth2/", BaseURL)
	client     *http.Client = &http.Client{}
	//Rune indices for lower/uppercase a/z, used for sgf string conversion
	rAL, rZL, rAU, rZU int = int('a'), int('z'), int('A'), int('Z')
)

//OGSApiError is returned on non-200 return codes from the online-go API.
type OGSApiError struct {
	Code int
	err  error
}

func (o *OGSApiError) Error() string {
	return o.err.Error()
}

//OauthResponse is returned by the oauth/token endpoint of OGS.
type OauthResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	TokenType        string `json:"token_type"`
	RefreshToken     string `json:"refresh_token"`
	RawScope         string `json:"scope"` //must still be parsed, not necessary at the moment
	Error            string `json:"error"` //will be empty string if all went well
	ErrorDescription string `json:"error_description"`
}

func (o *OauthResponse) GetError() string {
	if o.Error != "" {
		if o.ErrorDescription != "" {
			return o.ErrorDescription
		}
		//Oauth error with no description is likely misconfiguration
		panic(o.Error)
	}
	return ""
}

//UserInfo contains info about the logged in user after calling either Authenticate function.
type UserInfo struct {
	Authenticated bool
	Player        Player
	Oauth         OauthResponse
}

//Player information
type Player struct {
	ID         int64   `json:"id"`
	Username   string  `json:"username"`
	RawRanking float32 `json:"ranking"`
}

func (p Player) String() string {
	return fmt.Sprintf("%s (%s)", p.Username, p.Ranking())
}

func (p Player) Ranking() string {
	//adds 0.5 to round to nearest integer instead of rounding 1.9 to 1
	if p.RawRanking < 30 {
		return fmt.Sprintf("%d kyu", int(30-p.RawRanking+0.5))
	} else {
		return fmt.Sprintf("%d dan", int((p.RawRanking-30+0.5)+1))
	}
}

//Game information
type GameList struct {
	Games []GameListData `json:"results"`
}

//GameListData contains data from the current games endpoint. This does not contain game details;
//these are contained in BoardState and are only loaded when a game is selected.
type GameListData struct {
	ID        int64             `json:"id"`
	Name      string            `json:"name"`
	Width     int               `json:"width"`
	Height    int               `json:"height"`
	Players   map[string]Player `json:"players"`
	BlackLost bool              `json:"black_lost"`
	WhiteLost bool              `json:"white_lost"`
}

//BoardData is a currently unused struct used to unmarshal the older api/v1/games/<id> endpoint.
//It requires reconstructing the game state based on Moves, which is incomplete.
//Use BoardState and its related functions instead.
type BoardData struct {
	Width                 int               `json:"width"`
	Height                int               `json:"height"`
	InitialPlayer         string            `json:"initial_player"`
	Handicap              int               `json:"handicap"`
	FreeHandicapPlacement bool              `json:"free_handicap_placement"`
	InitialState          map[string]string `json:"initial_state"`
	Moves                 []BoardPos        `json:"moves"`
}

//BoardState unmarshals the termination-api/game endpoint. In OGS, Board is represented as a 2D array,
//containing values from 0 to 2 (0 = empty, 1 = black, 2 = white). Board is indexed as Board[y][x].
type BoardState struct {
	MoveNumber   int     `json:"move_number"`
	PlayerToMove int64   `json:"player_to_move"`
	Phase        string  `json:"phase"`
	Board        [][]int `json:"board"`
	Outcome      string  `json:"outcome"`
	Removal      [][]int `json:"removal"`
	LastMove     struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"last_move"`
}

func (b *BoardState) Finished() bool {
	return b.Phase == "finished"
}

func (b *BoardState) Height() int {
	return len(b.Board)
}

func (b *BoardState) Width() int {
	if b.Height() == 0 {
		return 0
	}
	return len(b.Board[0])
}

//ColorForMove(i) Get whether move i belongs to "black" or "white" (true = black, false = white)
//This is part of the unfinished BoardData struct and should not be relied on, but can serve
//as a starting point if you're considering using it.
func (b *BoardData) ColorForMove(i int) bool {
	initialPlayer := b.InitialPlayer == "black"
	// If an uneven number of moves is made, it's the other player's turn.
	// If free_handicap_placement is true, then the initial player gets g.Handicap moves
	// If free_handicap_placement is false, GameData.InitialState is populated instead
	// with a string describing coordinates in SGF notation.
	// A handicap of 1 essentially does nothing
	initialExtraMoves := 0
	if b.FreeHandicapPlacement {
		initialExtraMoves = b.Handicap - 1
	}
	if i <= initialExtraMoves {
		return initialPlayer
	}
	if (i-initialExtraMoves)%2 == 1 {
		//It is the other player's turn
		return !initialPlayer
	}
	return initialPlayer
}

type BoardPos struct {
	X int
	Y int
}

func (p *BoardPos) UnmarshalJSON(data []byte) error {
	var v []float64
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	p.X = int(v[0])
	p.Y = int(v[1])
	return nil
}

//Description returns a formatted description of the game. Do not rely on this string remaining stable;
//it is purely a utility function for termsuji and may change or be removed entirely.
func (g GameListData) Description() string {
	ended := ""
	if g.GameOver() {
		ended = " (ended)"
	}
	return fmt.Sprintf("%s (B) vs %s (W) (%dx%d)%s", g.Players["black"], g.Players["white"], g.Width, g.Height, ended)
}

//GameOver returns true if the game has ended, otherwise false.
func (g GameListData) GameOver() bool {
	//If a game is over, one of these will be false
	return !g.BlackLost || !g.WhiteLost
}

//GetGamesList returns a number of ongoing games you are actively participating in.
//Pagination is currently not implemented, so only the first few active games will be returned.
func GetGamesList() *GameList {
	var gamelist GameList
	var values url.Values = make(url.Values)
	values.Set("ended__isnull", "true")
	doGet(apiURL, "me/games", values, &gamelist)
	return &gamelist
}

//GetGameData returns metadata for the given game ID, if it is public. If it is not, BoardData will be uninitialized.
//For getting the actual contents of the board, consider using GetGameState.
func GetGameData(gameID int64) *BoardData {
	var board BoardData
	doGet(termApiURL, fmt.Sprintf("game/%d", gameID), nil, &board)
	return &board
}

//GetGameState returns the whole board's state and the last played move, among other things.
//This endpoint contains little to no other metadata.
func GetGameState(gameID int64) *BoardState {
	var board BoardState
	doGet(termApiURL, fmt.Sprintf("game/%d/state", gameID), nil, &board)
	return &board
}

//AuthenticateRefreshToken authenticates with a token from the user.
//Either this function or
func AuthenticateRefreshToken(refreshToken string) error {
	var oauthResponse OauthResponse
	var apiError *OGSApiError
	var values url.Values = make(url.Values)
	values.Set("client_id", OauthClientID)
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", refreshToken)
	err := doPostForm(oauthURL, "token/", values, &oauthResponse) //trailing slash to path is required!

	if errors.As(err, &apiError) && oauthResponse.Error != "" {
		return InvalidRefreshToken
	} else if err != nil {
		return err
	}
	AuthData.Oauth = oauthResponse
	getPlayerForAuth()
	return nil
}

func AuthenticatePassword(username, password string) error {
	var oauthResponse OauthResponse
	if username == "" || password == "" {
		return errors.New("Username/password required")
	}
	var values url.Values = make(url.Values)
	values.Set("client_id", OauthClientID)
	values.Set("grant_type", "password")
	values.Set("username", username)
	values.Set("password", password)
	err := doPostForm(oauthURL, "token/", values, &oauthResponse) //trailing slash to path is required!
	if oauthResponse.Error != "" {
		return errors.New(oauthResponse.GetError()) //may panic, depending on error
	} else if err != nil {
		return err
	}
	AuthData.Oauth = oauthResponse
	getPlayerForAuth()
	return nil
}

func getPlayerForAuth() {
	var me Player
	err := doGet(apiURL, "me", nil, &me)
	if err != nil {
		panic(err)
	}
	//if the call succeeds, user must be authenticated
	AuthData.Player = me
	AuthData.Authenticated = true
}

type OGSConfig struct {
	ChatAuth string `json:"chat_auth"`
}

//GetOGSConfig gets the ui/config endpoint from OGS.
//Only one parameter, chat_auth, is extracted for use with the realtime API.
func GetOGSConfig() *OGSConfig {
	o := &OGSConfig{}
	doGet(apiURL, "ui/config", nil, o)
	return o
}

func doPostForm(apiURL, apiPath string, values url.Values, unpack any) error {
	return handleRequest("POST", apiURL, apiPath, "", values, &unpack)
}

func doPostJSON(apiURL, apiPath string, jsonstr string, unpack any) error {
	return handleRequest("POST", apiURL, apiPath, jsonstr, nil, &unpack)
}

func doGet(apiURL, apiPath string, values url.Values, unpack any) error {
	return handleRequest("GET", apiURL, apiPath, "", values, &unpack)
}

func handleRequest(httpMethod string, apiURL, apiPath, jsonstr string, postValues url.Values, unpack *any) error {
	var r io.Reader
	if jsonstr != "" {
		r = strings.NewReader(jsonstr)
	}
	if httpMethod != "GET" && postValues != nil {
		r = strings.NewReader(postValues.Encode())
	}
	req, err := http.NewRequest(httpMethod, fmt.Sprintf("%s%s", apiURL, apiPath), r)
	if err != nil {
		return err
	}

	if postValues != nil {
		if httpMethod == "GET" {
			req.URL.RawQuery = postValues.Encode()
		} else {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}
	if AuthData.Oauth.AccessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AuthData.Oauth.AccessToken))
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(respData, &unpack)
	if resp.StatusCode != 200 {
		return &OGSApiError{Code: resp.StatusCode, err: errors.New(fmt.Sprintf("Error calling /%s: %s", apiPath, resp.Status))}
	}
	if err != nil {
		return err
	}
	return nil
}

//ConvertSGCoords turns an SGF coordinates string (2 letters for col+row) to a list of board positions.
//This doesn't contain any other context, like which player's turn it is.
//This function is currently not used for anything, but is left here as a reference.
func ConvertSGFCoords(sgf string) *[]BoardPos {
	if len(sgf)%2 == 1 {
		panic(fmt.Sprintf("invalid length for sgf coordinate string: %s", sgf))
	}

	var posList []BoardPos = make([]BoardPos, len(sgf)/2)
	for i := range posList {
		posList[i] = BoardPos{
			X: SGFInt(sgf[i*2]),
			Y: SGFInt(sgf[(i*2)+1]),
		}
	}
	return &posList
}

//SGFInt converts a sgf notation letter to integer, which is required for
//reading the initial state parameter of the legacy single game endpoint.
//This function is currently not used for anything, but is left here as a reference.
func SGFInt(r byte) int {
	rInt := int(r)
	switch {
	case rInt >= rAL && rInt <= rZL:
		//lowercase a corresponds to 0
		return rInt - rAL
	case rInt >= rAU && rInt <= rZU:
		//uppercase A comes after lowercase z
		return rInt - rAU + 26
	default:
		panic(fmt.Sprintf("invalid sgf coordinate rune: %c", r))
	}
}

//PosSGF converts a BoardPos x, y struct to SGF coordinate notation, which is required
//for posting moves to the realtime API.
//The SGF notation has two letters per coordinate, where "aa" is the upper left corner,
//"ba" is one stone to the right of that, "bc" is two stones below "ba", etc.
func PosSGF(p BoardPos) string {
	if p.X == -1 && p.Y == -1 {
		//Not official, but used by OGS
		return ".."
	}
	//Uppercase letters aren't implemented at the moment, so it's only up to 26x26 boards
	return fmt.Sprintf("%c%c", rune(rAL+p.X), rune(rAL+p.Y))
}
