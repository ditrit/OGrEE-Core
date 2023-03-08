package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	flag "github.com/spf13/pflag"
)

type Config struct {
	Verbose      string
	UnityURL     string
	UnityTimeout string
	APIURL       string
	APIKEY       string
	ConfigPath   string
	HistPath     string
	Analyser     bool
	Script       string
	Drawable     []string
	DrawableJson map[string]string
	DrawLimit    int
	Updates      []string
	User         string
}

func defaultConfig() Config {
	return Config{
		Verbose:      "ERROR",
		UnityURL:     "",
		UnityTimeout: "10ms",
		APIURL:       "",
		APIKEY:       "",
		ConfigPath:   "./config.toml",
		HistPath:     "./.history",
		Analyser:     true,
		Script:       "",
		Drawable:     []string{"all"},
		DrawableJson: map[string]string{},
		DrawLimit:    50,
		Updates:      []string{"all"},
		User:         "",
	}
}

func ReadConfig() *Config {
	conf := defaultConfig()
	flag.StringVarP(&conf.ConfigPath, "conf_path", "c", conf.ConfigPath,
		"Indicate the location of the Shell's config file")
	flag.Parse()
	configBytes, err := os.ReadFile(conf.ConfigPath)
	if err != nil {
		fmt.Println("Cannot read config file", conf.ConfigPath, ":", err.Error())
		fmt.Println("Please ensure that you have a properly formatted config file saved as 'config.toml' in the current directory")
		fmt.Println("\n\nFor more details please refer to: https://ogree.ditrit.io/htmls/programming.html")
		fmt.Println("View an environment file example here: https://ogree.ditrit.io/htmls/clienv.html")
	}
	_, err = toml.Decode(string(configBytes), &conf)
	if err != nil {
		println("Error reading config :", err.Error())
	}
	flag.StringVarP(&conf.Verbose, "verbose", "v", conf.Verbose,
		"Indicates level of debugging messages."+
			"The levels are of in ascending order:"+
			"{NONE,ERROR,WARNING,INFO,DEBUG}.")
	flag.StringVarP(&conf.UnityURL, "unity_url", "u", conf.UnityURL, "Unity URL")
	flag.StringVarP(&conf.APIURL, "api_url", "a", "", "API URL")
	flag.StringVarP(&conf.APIKEY, "api_key", "k", conf.APIKEY, "Indicate the key of the API")
	flag.StringVarP(&conf.HistPath, "history_path", "h", conf.HistPath,
		"Indicate the location of the Shell's history file")
	flag.BoolVarP(&conf.Analyser, "analyser", "s", conf.Analyser, "Dictate if the Shell shall"+
		" use the Static Analyser before script execution")
	flag.StringVarP(&conf.Script, "file", "f", conf.Script, "Launch the shell as an interpreter "+
		" by only executing an OCLI script file")
	flag.Parse()
	return &conf
}

func UpdateConfigFile(conf *Config) error {
	configFile, err := os.Create(conf.ConfigPath)
	if err != nil {
		return fmt.Errorf("cannot open config file to edit user and key")
	}
	err = toml.NewEncoder(configFile).Encode(conf)
	if err != nil {
		panic("invalid config : " + err.Error())
	}
	return nil
}
