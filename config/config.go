package config

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"

	"github.com/adrg/xdg"
)

var (
	cfgFile       = "termsuji/config.json"
	authFile      = "termsuji/auth.json"
	defaultConfig = Config{
		BoardColor:    220,
		BoardColorAlt: 221,
		BlackColor:    233,
		BlackColorAlt: 235,
		WhiteColor:    255,
		WhiteColorAlt: 254,
		CursorColor:   2,
	}
)

type Config struct {
	BoardColor    int `json:"board_clr1"`
	BoardColorAlt int `json:"board_clr2"`
	BlackColor    int `json:"black_clr1"`
	BlackColorAlt int `json:"black_clr2"`
	WhiteColor    int `json:"white_clr1"`
	WhiteColorAlt int `json:"white_clr2"`
	CursorColor   int `json:"cursor_clr"`
}

func InitConfig() *Config {
	config := defaultConfig
	absPath, err := xdg.SearchConfigFile(cfgFile)
	if err == nil {
		readCfgFile(absPath, &config)
	}
	return &config
}

func (c *Config) Save() {
	absPath, err := xdg.ConfigFile(cfgFile)
	if err != nil {
		panic(err)
	}
	saveCfgFile(absPath, c, 0664)
}

type AuthData struct {
	Username string `json:"username"`
	UserID   int64  `json:"id"`
	Tokens   struct {
		Refresh string `json:"refresh"`
	} `json:"tokens"`
}

func InitAuthData() *AuthData {
	authData := AuthData{}
	absPath, err := xdg.SearchStateFile(authFile)
	if err == nil {
		readCfgFile(absPath, &authData)
	}
	return &authData
}

func (a *AuthData) Save() {
	absPath, err := xdg.StateFile(authFile)
	if err != nil {
		panic(err)
	}
	saveCfgFile(absPath, a, 0600)
}

func saveCfgFile(filePath string, a interface{}, perm fs.FileMode) {
	jsonData, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(filePath, jsonData, perm)
	if err != nil {
		panic(err)
	}
}

func readCfgFile(filePath string, a interface{}) *interface{} {
	configReader, err := ioutil.ReadFile(filePath)
	if err == nil {
		err = json.Unmarshal(configReader, &a)
		if err != nil {
			panic(err)
		}
	}
	return &a
}
