package controllers

import (
	"cli/logger"
	"cli/models"
	"cli/readline"
	"errors"
	"fmt"
	"net"
)

const defaultOgree3DURL = "localhost:5500"

var Ogree3D Ogree3DPort = &ogree3DPortImpl{
	connection: models.Ogree3DConnection{},
}

type Ogree3DPort interface {
	URL() string
	SetURL(url string) error
	SetDefaultURL()

	Connect(url string, rl *readline.Instance) error
	Disconnect()
	IsConnected() bool

	// Sends a message to OGrEE-3D
	//
	// If there isn't a connection established, tries to establish the connection first
	Inform(caller string, entity int, data map[string]interface{}) error
	// Sends a message to OGrEE-3D if there is a connection established,
	// otherwise does nothing
	InformOptional(caller string, entity int, data map[string]interface{}) error
}

type ogree3DPortImpl struct {
	url        string
	connection models.Ogree3DConnection
}

func (ogree3D *ogree3DPortImpl) URL() string {
	return ogree3D.url
}

func (ogree3D *ogree3DPortImpl) SetURL(url string) error {
	if url == "" {
		ogree3D.SetDefaultURL()
		return nil
	}

	_, _, err := net.SplitHostPort(url)
	if err != nil {
		return fmt.Errorf("OGrEE-3D URL is not valid: %s", url)
	}

	ogree3D.url = url

	return nil
}

func (ogree3D *ogree3DPortImpl) SetDefaultURL() {
	if ogree3D.url != defaultOgree3DURL {
		msg := fmt.Sprintf("Falling back to default OGrEE-3D URL: %s", defaultOgree3DURL)
		fmt.Println(msg)
		logger.GetInfoLogger().Println(msg)
		ogree3D.url = defaultOgree3DURL
	}
}

func (ogree3D *ogree3DPortImpl) Connect(url string, rl *readline.Instance) error {
	if ogree3D.connection.IsConnected() {
		if url == "" || url == ogree3D.url {
			return fmt.Errorf("already connected to OGrEE-3D url: %s", ogree3D.url)
		} else {
			ogree3D.connection.Disconnect()
		}
	}

	if url == "" {
		fmt.Printf("Using OGrEE-3D url: %s\n", ogree3D.url)
	} else {
		err := ogree3D.SetURL(url)
		if err != nil {
			return err
		}
	}

	return ogree3D.initCommunication(rl)
}

// Tries to establish a connection with OGrEE-3D and, if possible,
// starts a go routine for receiving messages from it
func (ogree3D *ogree3DPortImpl) initCommunication(rl *readline.Instance) error {
	errConnect := ogree3D.connection.Connect(ogree3D.url, State.Timeout)
	if errConnect != nil {
		return ErrorWithInternalError{
			UserError:     errors.New("OGrEE-3D is not reachable"),
			InternalError: errConnect,
		}
	}

	errLogin := ogree3D.login(State.APIURL, GetKey(), State.DebugLvl)
	if errLogin != nil {
		return ErrorWithInternalError{
			UserError:     errors.New("OGrEE-3D login not possible"),
			InternalError: errLogin,
		}
	}

	fmt.Println("Established connection with OGrEE-3D!")

	go ogree3D.connection.ReceiveLoop(rl)

	return nil
}

// Transfer login apiKey for the OGrEE-3D to communicate with the API
func (ogree3D *ogree3DPortImpl) login(apiURL, apiToken string, debugLevel int) error {
	data := map[string]interface{}{"api_url": apiURL, "api_token": apiToken}
	req := map[string]interface{}{"type": "login", "data": data}

	return ogree3D.connection.Send(req, debugLevel)
}

func (ogree3D *ogree3DPortImpl) Disconnect() {
	ogree3D.connection.Disconnect()
}

func (ogree3D *ogree3DPortImpl) IsConnected() bool {
	return ogree3D.connection.IsConnected()
}

// Sends a message to OGrEE-3D
//
// If there isn't a connection established, tries to establish the connection first
func (ogree3D *ogree3DPortImpl) Inform(caller string, entity int, data map[string]interface{}) error {
	if !ogree3D.connection.IsConnected() {
		fmt.Println("Connecting to OGrEE-3D")
		err := Connect3D("")
		if err != nil {
			return err
		}
	}

	return ogree3D.InformOptional(caller, entity, data)
}

// Sends a message to OGrEE-3D if there is a connection established,
// otherwise does nothing
func (ogree3D *ogree3DPortImpl) InformOptional(caller string, entity int, data map[string]interface{}) error {
	if ogree3D.connection.IsConnected() {
		if entity > -1 && entity <= models.CORRIDOR {
			data = GenerateFilteredJson(data)
		}
		if State.DebugLvl > INFO {
			println("DEBUG VIEW THE JSON")
			Disp(data)
		}

		e := ogree3D.connection.Send(data, State.DebugLvl)
		if e != nil {
			logger.GetWarningLogger().Println("Unable to contact Unity Client @" + caller)
			if State.DebugLvl > 1 {
				fmt.Println("Error while updating Unity: ", e.Error())
			}
			return fmt.Errorf("error while contacting unity : %s", e.Error())
		}
	}
	return nil
}

func Connect3D(url string) error {
	return Ogree3D.Connect(url, *State.Terminal)
}

func Disconnect3D() {
	Ogree3D.InformOptional("Disconnect3d", -1, map[string]interface{}{"type": "logout", "data": ""})
	Ogree3D.Disconnect()
}

// This func is used for when the user wants to filter certain
// attributes from being sent/displayed to Unity viewer client
func GenerateFilteredJson(x map[string]interface{}) map[string]interface{} {
	ans := map[string]interface{}{}
	attrs := map[string]interface{}{}
	if catInf, ok := x["category"]; ok {
		if cat, ok := catInf.(string); ok {
			if models.EntityStrToInt(cat) != -1 {

				//Start the filtration
				for i := range x {
					if i == "attributes" {
						for idx := range x[i].(map[string]interface{}) {
							if IsCategoryAttrDrawable(x["category"].(string), idx) {
								attrs[idx] = x[i].(map[string]interface{})[idx]
							}
						}
					} else {
						if IsCategoryAttrDrawable(x["category"].(string), i) {
							ans[i] = x[i]
						}
					}
				}
				if len(attrs) > 0 {
					ans["attributes"] = attrs
				}
				return ans
			}
		}
	}
	return x //Nothing will be filtered
}

func IsInObjForUnity(entityStr string) bool {
	entInt := models.EntityStrToInt(entityStr)
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
