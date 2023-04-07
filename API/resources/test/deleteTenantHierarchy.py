#!/usr/bin/env python3
import requests, argparse, os, json, sys

#FLAG
res = True

#CONSTANT
expected = 404
PIDS={  "tenantID":None, "siteID":None, "buildingID":None,
        "roomID":None,"acID":None,"panelID":None,
  "cabinetID":None, "groupID":None, "corridorID":None,
  "room-sensorID":None,"rack-sensorID":None,
  "device-sensorID":None,
  "rackID":None, "deviceID":None, 
    "room-templateID":None, "obj-templateID":None
     }

#Function
def checkResponse(code, entity):
    global res
    if code == expected:
        print(entity+" Empty: Successful")
    else:
        res = False
        print(entity+" Empty: Fail")

def writeEnv():
    with open('.localenv', 'r') as file:
        data = file.read() # use `json.loads` to do the reverse
        x = json.loads(data)
        for i in x:
            PIDS[i] = x[i]

def delObjHelper(uri):
    return requests.request("DELETE", uri, 
                              headers=headers, data=payload)

def delObj(uri, obj):
    response = delObjHelper(uri)
    if response.status_code == 204:
        print(obj+" successfully deleted")
    else:
        print("Failed to delete "+obj)


#Setup Arg and ENV
parser = argparse.ArgumentParser(description=
        'Delete Tenant Hierarchy Test Script')
parser.add_argument('--url', 
                    help="""Specify which API URL to use""")
writeEnv()
tid = PIDS['tenantID']



#Setup API URL
args = vars(parser.parse_args())
if ('url' not in args or args['url'] == None):
    print('API URL not specified... using default URL')
    url = "http://localhost:3001/api"
else:
    url = args['url']


entRange=entRange=["tenant","site","building","room","ac",
  "panel","cabinet", "group", "corridor",
  "room-sensor","rack-sensor","device-sensor",
          "rack","device","room-template","obj-template"]
payload={}
headers = {
  'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjAsIkN1c3RvbWVyIjoib2dyZWUifQ.3A0ZI7oSXxhZ8CGhJORYRXIm3MKroo_9LM939dgGHIo'
}


#START
delObj(url+"/tenants/"+tid, "Tenant")
delObj(url+"/room-templates/"+PIDS["room-templateID"], "Room-Template")
delObj(url+"/obj-templates/"+PIDS["obj-templateID"], "Obj-Template")
delObj(url+"/groups/"+PIDS["groupID"], "Group")



for i in entRange:
  URL = None
  ID = PIDS[i+"ID"]

  #Sensors have a different URL
  if i.find("-sensor") != -1:
      URL = url+"/sensors/"+ID
  else:
      URL = url+"/"+i+"s"+"/"+ID

  print("Now testing: ", URL)
  response = requests.request("GET", URL, headers=headers, data=payload)
  checkResponse(response.status_code, i)

  #print(response.text)

#RETURN VALUE
if res != True:
    sys.exit(-1)

#else success
sys.exit()