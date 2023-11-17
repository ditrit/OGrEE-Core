package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"cli/utils"

	"github.com/BurntSushi/toml"
	flag "github.com/spf13/pflag"
)

type globalConfig struct {
	Conf Config `toml:"OGrEE-CLI"`
}

type Vardef struct {
	Name  string
	Value any
}

type Config struct {
	Verbose      string
	APIURL       string
	UnityURL     string
	UnityTimeout string
	ConfigPath   string
	HistPath     string
	Script       string
	Drawable     []string
	DrawableJson map[string]string
	DrawLimit    int
	Updates      []string
	User         string
	Password     string
	Variables    []Vardef
}

// Used for parsing (via JSON) into conf after parsing TOML
// since an object can only be decoded by TOML once
type ArgStruct struct {
	ConfigPath string `json:",omitempty"`
	Verbose    string `json:",omitempty"`
	UnityURL   string `json:",omitempty"`
	APIURL     string `json:",omitempty"`
	HistPath   string `json:",omitempty"`
	Script     string `json:",omitempty"`
	User       string `json:",omitempty"`
	Password   string `json:",omitempty"`
}

func defaultConfig() Config {
	return Config{
		Verbose:      "ERROR",
		APIURL:       "",
		UnityURL:     "",
		UnityTimeout: "10ms",
		ConfigPath:   utils.ExeDir() + "/../config.toml",
		HistPath:     "./.history",
		Script:       "",
		Drawable:     []string{"all"},
		DrawableJson: map[string]string{},
		DrawLimit:    50,
		Updates:      []string{"all"},
		User:         "",
		Password:     "",
		Variables:    []Vardef{},
	}
}

func ReadConfig() (*Config, error) {
	globalConf := globalConfig{
		Conf: defaultConfig(),
	}
	args := ArgStruct{}
	conf := &globalConf.Conf

	flag.StringVarP(&args.ConfigPath, "conf_path", "c", conf.ConfigPath,
		"Indicate the location of the Shell's config file")
	flag.StringVarP(&args.Verbose, "verbose", "v", conf.Verbose,
		"Indicates level of debugging messages."+
			"The levels are of in ascending order:"+
			"{NONE,ERROR,WARNING,INFO,DEBUG}.")
	flag.StringVarP(&args.UnityURL, "unity_url", "u", conf.UnityURL, "Unity URL")
	flag.StringVarP(&args.APIURL, "api_url", "a", conf.APIURL, "API URL")
	flag.StringVarP(&args.HistPath, "history_path", "h", conf.HistPath,
		"Indicate the location of the Shell's history file")
	flag.StringVarP(&args.Script, "file", "f", conf.Script, "Launch the shell as an interpreter "+
		" by only executing an OCLI script file")
	flag.StringVarP(&args.User, "user", "", conf.User, "User email")
	flag.StringVarP(&args.Password, "password", "", conf.Password, "Password")
	flag.Parse()

	configBytes, err := os.ReadFile(args.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf(
			`cannot read config file %s : %s\n
			Please ensure that you have a properly formatted config file saved as 'config.toml' in the parent directory
			For more details please refer to: https://github.com/ditrit/OGrEE-Core/blob/main/README.md`,
			args.ConfigPath, err.Error())
	}
	_, err = toml.Decode(string(configBytes), &globalConf)
	if err != nil {
		return nil, fmt.Errorf("error reading config : %s", err.Error())
	}
	conf.Password = ""

	argBytes, _ := json.Marshal(args)
	json.Unmarshal(argBytes, &conf)

	conf.ConfigPath, _ = filepath.Abs(conf.ConfigPath)
	conf.HistPath, _ = filepath.Abs(conf.HistPath)
	return conf, nil
}
