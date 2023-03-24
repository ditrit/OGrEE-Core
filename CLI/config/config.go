package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	flag "github.com/spf13/pflag"
)

type globalConfig struct {
	Conf Config `toml:"OGrEE-CLI"`
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
	APIKEY       string
}

func defaultConfig() Config {
	return Config{
		Verbose:      "ERROR",
		APIURL:       "",
		UnityURL:     "",
		UnityTimeout: "10ms",
		ConfigPath:   "../config.toml",
		HistPath:     "./.history",
		Script:       "",
		Drawable:     []string{"all"},
		DrawableJson: map[string]string{},
		DrawLimit:    50,
		Updates:      []string{"all"},
		User:         "",
		APIKEY:       "",
	}
}

func ReadConfig() *Config {
	globalConf := globalConfig{
		Conf: defaultConfig(),
	}
	conf := &globalConf.Conf
	flag.StringVarP(&conf.ConfigPath, "conf_path", "c", conf.ConfigPath,
		"Indicate the location of the Shell's config file")
	flag.Parse()
	configBytes, err := os.ReadFile(conf.ConfigPath)
	if err != nil {
		fmt.Println("Cannot read config file", conf.ConfigPath, ":", err.Error())
		fmt.Println("Please ensure that you have a properly formatted config file saved as 'config.toml' in the parent directory")
		fmt.Println("For more details please refer to: https://github.com/ditrit/OGrEE-Core/blob/main/README.md")
	}
	_, err = toml.Decode(string(configBytes), &globalConf)
	if err != nil {
		println("Error reading config :", err.Error())
	}
	flag.StringVarP(&conf.Verbose, "verbose", "v", conf.Verbose,
		"Indicates level of debugging messages."+
			"The levels are of in ascending order:"+
			"{NONE,ERROR,WARNING,INFO,DEBUG}.")
	flag.StringVarP(&conf.UnityURL, "unity_url", "u", conf.UnityURL, "Unity URL")
	flag.StringVarP(&conf.APIURL, "api_url", "a", conf.APIURL, "API URL")
	flag.StringVarP(&conf.APIKEY, "api_key", "k", conf.APIKEY, "Indicate the key of the API")
	flag.StringVarP(&conf.HistPath, "history_path", "h", conf.HistPath,
		"Indicate the location of the Shell's history file")
	flag.StringVarP(&conf.Script, "file", "f", conf.Script, "Launch the shell as an interpreter "+
		" by only executing an OCLI script file")
	flag.Parse()
	return conf
}
