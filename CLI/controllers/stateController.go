package controllers

import (
	"cli/logger"
	"cli/readline"
	"fmt"
	"net"
	"time"
)

var BuildTime string
var BuildHash string
var BuildTree string
var GitCommitDate string
var State ShellState

type User struct {
	Email string
	ID    string
}

const defaultOgree3DURL = "localhost:5500"

type ShellState struct {
	Prompt             string
	BlankPrompt        string
	Customer           string //Tenant name
	CurrPath           string
	PrevPath           string
	ClipBoard          []string
	Hierarchy          *HierarchyNode
	ConfigPath         string //Holds file path of '.env'
	HistoryFilePath    string //Holds file path of '.history'
	Ogree3DURL         string
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
}

func (state ShellState) SetOgree3DURL(ogree3DURL string) error {
	if ogree3DURL == "" {
		state.SetDefaultOgree3DURL()
		return nil
	}

	_, _, err := net.SplitHostPort(ogree3DURL)
	if err != nil {
		return fmt.Errorf("OGrEE-3D URL is not valid: %s", ogree3DURL)
	}

	State.Ogree3DURL = ogree3DURL

	return nil
}

func (state ShellState) SetDefaultOgree3DURL() {
	if State.Ogree3DURL != defaultOgree3DURL {
		msg := fmt.Sprintf("Falling back to default OGrEE-3D URL: %s", defaultOgree3DURL)
		fmt.Println(msg)
		logger.GetInfoLogger().Println(msg)
		State.Ogree3DURL = defaultOgree3DURL
	}
}

func IsInObjForUnity(x string) bool {
	entInt := EntityStrToInt(x)
	if entInt != -1 {

		for idx := range State.ObjsForUnity {
			if State.ObjsForUnity[idx] == entInt {
				return true
			}
		}
	}
	return false
}

func IsDrawableEntity(x string) bool {
	entInt := EntityStrToInt(x)

	for idx := range State.DrawableObjs {
		if State.DrawableObjs[idx] == entInt {
			return true
		}
	}
	return false
}

func GetKey() string {
	return State.APIKEY
}
