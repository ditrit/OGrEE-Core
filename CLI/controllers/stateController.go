package controllers

import (
	"cli/readline"
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

func IsInObjForUnity(entityStr string) bool {
	entInt := EntityStrToInt(entityStr)
	return IsEntityTypeForOGrEE3D(entInt)
}

func IsEntityTypeForOGrEE3D(entityType int) bool {
	if entityType != -1 {
		for idx := range State.ObjsForUnity {
			if State.ObjsForUnity[idx] == entityType {
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
