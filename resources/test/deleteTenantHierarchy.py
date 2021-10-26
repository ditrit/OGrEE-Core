#!/usr/bin/env python3
import requests, argparse, os, json, sys

#FLAG
res = True

#CONSTANT
expected = 404
PIDS={  "tenantID":None, "siteID":None, "buildingID":None,
        "roomID":None, "rackID":None, "deviceID":None, 
        "subdeviceID":None, "subdevice1ID":None, "room-templateID":None,
        "obj-templateID":None
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

def delObj(uri):
    return requests.request("DELETE", uri, 
                              headers=headers, data=payload)


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
    url = "http://localhost:27020/api"
else:
    url = args['url']


entRange=entRange=["tenant","site","building","room",
          "rack","device","room-template","obj-template"]
payload={}
headers = {
  'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjY2NDA0NjEyNzM0MjQxOTk2OX0.cB1VkYQLlXCatzMiEWGFfJKKx9h8Vsr2vdlylNMe7hs'
}


#START
response = delObj(url+"/tenants/"+tid)
if response.status_code == 204:
    print("Tenant successfully deleted")
else:
    print("Failed to delete Tenant")


response = delObj(url+"/room-templates/"+PIDS["room-templateID"])
if response.status_code == 204:
    print("Room Template successfully deleted")
else:
    print("Failed to delete Room Template")


response = delObj(url+"/obj-templates/"+PIDS["obj-templateID"])
if response.status_code == 204:
    print("Obj Template successfully deleted")
else:
    print("Failed to delete Obj Template")



for i in entRange:
  ID = PIDS[i+"ID"]
  print("Now testing: ", url+"/"+i+"s"+"/"+ID)
  response = requests.request("GET", url+"/"+i+"s"+"/"+ID, headers=headers, data=payload)
  checkResponse(response.status_code, i)

  #print(response.text)

#RETURN VALUE
if res != True:
    sys.exit(-1)

#else success
sys.exit(1)