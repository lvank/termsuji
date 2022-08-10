# Termsuji

termsuji is an application to play Go in your terminal. It is limited in features and scope, but you can play and finish games in progress on it. It is on github as a reference implementation, and not a reliable or stable package.

The *api* package can be used as a starting point to work with the online-go.com REST and realtime APIs. It only exports a limited part of the API and it may change without notice.

If you want to build yourself (or if your architecture isn't listed), [download/install Go 1.18 or higher](https://go.dev/dl), download and extract the source code, register an Oauth application at https://online-go.com/oauth2/applications/ (this requires an online-go.com account), set the client type to "Public" and the grant type to "resource owner password based", place the client ID in `api/client_id.txt` without any whitespace, and run/build the application with `go run .` or `go build .` in the source code directory

![termsuji3](https://user-images.githubusercontent.com/110688516/183505721-6e50c05d-2572-4bb0-a06d-eae3006414a3.png)
![termsuji4](https://user-images.githubusercontent.com/110688516/183740301-19c66b74-d0ba-4fc2-a380-c9a3c08632e1.png)

## Configuration

There's a themes option in-application with some preset themes, but you can get more detailed configuration by editing the configuration file.
The application stores a configuration file in $XDG_CONFIG_HOME/termsuji/config.json (or C:/Users/YourUsername/AppData/Roaming/termsuji/config.json on Windows) with the following configurable values.
To find out "Unicode code points", you can visit https://unicode-table.com/, look up a symbol you want to use and copy the number from "HTML code" (or for the technically inclined, convert the Unicode number from hex to int).
To find out colour numbers, refer to the bottom left numbers on https://upload.wikimedia.org/wikipedia/commons/1/15/Xterm_256color_chart.svg
To reset your default settings, just delete the configuration file and it'll be regenerated with the defaults.

(Note: for symbols, you can't use 1-31 and 127-159.)

```
{"theme": {
  "draw_stone_bg": If true, will draw a stone with the black/black_alt colours. If false, draws the black/white symbols instead. Default true.
  "draw_cursor_bg": If true, will draw the currently selected square with the cursor_bg colour. If false, draws the cursor symbol instead. Default false.
  "draw_last_played_bg": If true, will draw the last played stone with the last_played_bg colour. If false, draws the last_played symbol instead. Default false.
  "fullwidth_letters": If true, will draw letter coordinates as fullwidth Japanese characters, occupying two spaces. Default false.
  "colors": {
    "board":             Go board background colour
    "board_alt":         Go board background colour (for alternating squares)
    "black":             Black player stone colour
    "black_alt":         Black player stone colour (for alternating squares)
    "white":             White player stone colour
    "white_alt":         White player stone colour (for alternating squares)
    "cursor_fg":         Cursor foreground colour (if no stone is selected)
    "cursor_bg":         Cursor background colour (if draw_stone_bg is true)
    "last_played_bg":    Last played stone background colour (if draw_last_played_bg is true)
  },
  "symbols": {
    "black":       Unicode code point for black stones. Default 32 (a blank space)
    "white":       Unicode code point for white stones. Default 32 (a blank space)
    "board":       Unicode code point for the board itself. Default 32 (a blank space)
    "cursor":      Unicode code point for the cursor. Default 88, or 'X'
    "last_played": Unicode code point to mark the last played stone. Default 47, or '/'
  }
}}
```