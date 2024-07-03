package controllers

import (
	"cli/commands"
	l "cli/logger"
	"cli/models"
	"cli/readline"
	"cli/utils"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var BuildTime string
var BuildHash string
var BuildTree string
var GitCommitDate string

var State ShellState

type ShellState struct {
	Prompt             string
	BlankPrompt        string
	Customer           string //Tenant name
	CurrPath           string
	CurrDomain         string
	PrevPath           string
	ClipBoard          []string
	Hierarchy          *HierarchyNode
	ConfigPath         string //Holds file path of '.env'
	HistoryFilePath    string //Holds file path of '.history'
	User               User
	APIURL             string
	APIKEY             string
	FilterDisplay      bool  //Set whether or not to send attributes to unity
	ObjsForUnity       []int //Deciding what objects should be sent to unity
	DrawThreshold      int   //Number of objects to be sent at a time to unity
	DrawableObjs       []int //Indicate which objs drawable in unity
	DrawableJsons      map[string]map[string]interface{}
	DebugLvl           int
	Terminal           **readline.Instance
	Timeout            time.Duration
	DynamicSymbolTable map[string]interface{}
	FuncTable          map[string]interface{}
	DryRun             bool
	DryRunErrors       []error
}

func Clear() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Printf("\033[2J\033[H")
	}
}

// Function is an abstraction of a normal exit
func Exit() {
	//writeHistoryOnExit(&State.sessionBuffer)
	//runtime.Goexit()
	os.Exit(0)
}

func Help(entry string) {
	var path string
	entry = strings.TrimSpace(entry)
	switch entry {
	case "ls", "pwd", "print", "printf", "cd", "tree", "get", "clear",
		"lsog", "grep", "for", "while", "if", "env",
		"cmds", "var", "unset", "selection", commands.Connect3D, commands.Disconnect3D, "camera", "ui", "hc", "drawable",
		"link", "unlink", "draw", "getu", "getslot", "undraw",
		"lsenterprise", commands.Cp:
		path = "./other/man/" + entry + ".txt"

	case ">":
		path = "./other/man/focus.txt"

	case "+":
		path = "./other/man/plus.txt"

	case "=":
		path = "./other/man/equal.txt"

	case "-":
		path = "./other/man/minus.txt"

	case ".template":
		path = "./other/man/template.txt"

	case ".cmds":
		path = "./other/man/cmds.txt"

	case ".var":
		path = "./other/man/var.txt"

	case "lsobj", "lsten", "lssite", commands.LsBuilding, "lsroom", "lsrack",
		"lsdev", "lsac", "lscorridor", "lspanel", "lscabinet":
		path = "./other/man/lsobj.txt"

	default:
		path = "./other/man/default.txt"
	}
	text, e := os.ReadFile(utils.ExeDir() + "/" + path)
	if e != nil {
		println("Manual Page not found!")
	} else {
		println(string(text))
	}

}

func ShowClipBoard() []string {
	if State.ClipBoard != nil {
		for _, k := range State.ClipBoard {
			println(k)
		}
		return State.ClipBoard
	}
	return nil
}

func (controller Controller) SetClipBoard(x []string) ([]string, error) {
	State.ClipBoard = x
	var data map[string]interface{}

	if len(x) == 0 { //This means deselect
		data = map[string]interface{}{"type": "select", "data": "[]"}
		err := controller.Ogree3D.InformOptional("SetClipBoard", -1, data)
		if err != nil {
			return nil, fmt.Errorf("cannot reset clipboard : %s", err.Error())
		}
	} else {
		//Verify paths
		arr := []string{}
		for _, val := range x {
			obj, err := controller.GetObject(val)
			if err != nil {
				return nil, err
			}
			id, ok := obj["id"].(string)
			if ok {
				arr = append(arr, id)
			}
		}
		serialArr := "[\"" + strings.Join(arr, "\",\"") + "\"]"
		data = map[string]interface{}{"type": "select", "data": serialArr}
		err := controller.Ogree3D.InformOptional("SetClipBoard", -1, data)
		if err != nil {
			return nil, fmt.Errorf("cannot set clipboard : %s", err.Error())
		}
	}
	return State.ClipBoard, nil
}

// Displays environment variable values
// and user defined variables and funcs
func Env(userVars, userFuncs map[string]interface{}) {
	fmt.Println("Filter: ", State.FilterDisplay)
	fmt.Println()
	fmt.Println("Objects Unity shall be informed of upon update:")
	for _, k := range State.ObjsForUnity {
		fmt.Println(k)
	}
	fmt.Println()
	fmt.Println("Objects Unity shall draw:")
	for _, k := range State.DrawableObjs {
		fmt.Println(models.EntityToString(k))
	}

	fmt.Println()
	fmt.Println("Currently defined user variables:")
	for name, k := range userVars {
		if k != nil {
			fmt.Println("Name:", name, "  Value: ", k)
		}

	}

	fmt.Println()
	fmt.Println("Currently defined user functions:")
	for name := range userFuncs {
		fmt.Println("Name:", name)
	}
}

func SetEnv(arg string, val interface{}) {
	switch arg {
	case "Filter":
		if _, ok := val.(bool); !ok {
			msg := "Can only assign bool values for " + arg + " Env Var"
			l.GetWarningLogger().Println(msg)
			if State.DebugLvl > 0 {
				println(msg)
			}
		} else {
			if arg == "Filter" {
				State.FilterDisplay = val.(bool)
			}

			println(arg + " Display Environment variable set")
		}

	default:
		println(arg + " is not an environment variable")
	}
}

func LSOG() error {
	fmt.Println("********************************************")
	fmt.Println("OGREE Shell Information")
	fmt.Println("********************************************")

	fmt.Println("USER EMAIL:", State.User.Email)
	fmt.Println("API URL:", State.APIURL+"/api/")
	fmt.Println("OGrEE-3D URL:", Ogree3D.URL())
	fmt.Println("OGrEE-3D connected: ", Ogree3D.IsConnected())
	fmt.Println("BUILD DATE:", BuildTime)
	fmt.Println("BUILD TREE:", BuildTree)
	fmt.Println("BUILD HASH:", BuildHash)
	fmt.Println("COMMIT DATE: ", GitCommitDate)
	fmt.Println("CONFIG FILE PATH: ", State.ConfigPath)
	fmt.Println("LOG PATH:", "./log.txt")
	fmt.Println("HISTORY FILE PATH:", State.HistoryFilePath)
	fmt.Println("DEBUG LEVEL: ", State.DebugLvl)

	fmt.Printf("\n\n")
	fmt.Println("********************************************")
	fmt.Println("API Information")
	fmt.Println("********************************************")

	//Get API Information here
	resp, err := API.Request("GET", "/api/version", nil, http.StatusOK)
	if err != nil {
		return err
	}
	apiInfo, ok := resp.Body["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid response from API on GET /api/version")
	}
	fmt.Println("BUILD DATE:", apiInfo["BuildDate"])
	fmt.Println("BUILD TREE:", apiInfo["BuildTree"])
	fmt.Println("BUILD HASH:", apiInfo["BuildHash"])
	fmt.Println("COMMIT DATE: ", apiInfo["CommitDate"])
	fmt.Println("CUSTOMER: ", apiInfo["Customer"])
	return nil
}
