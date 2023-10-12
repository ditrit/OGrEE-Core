package handlers

import (
	"kube-admin/models"
	"kube-admin/services/cmd"
	"kube-admin/services/helm"
	"kube-admin/services/k8s"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// swagger:operation GET /tenants Tenants GetTenants
// Get Tenants on the kubernetes
// ---
// produces:
// - application/json
// security:
//   - Bearer: []
// responses:
//     '200':
//         description: ok
//     '400':
//         description: Bad request
//     '500':
//         description: Internal server error
func GetTenants(c *gin.Context){

	ns,err := k8s.GetNamespace()
	if err != nil{
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"tenants": ns})
}
// swagger:operation GET /tenants/{tenant} Tenants GetTenants
// Get Pods on  kubernetes namespace
// ---
// produces:
// - application/json
// parameters:
//   - name: tenant
//     in: path
//     description: tenant looking for
//     required: true
//     type: string
// security:
//   - Bearer: []
// responses:
//     '200':
//         description: ok
//     '400':
//         description: Bad request
//     '500':
//         description: Internal server error
func GetTenantPodsInfo(c *gin.Context) {
	name := c.Param("name")
	name = "ogree-"+name
	if response, err := k8s.GetPods(name); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	} else {
		c.IndentedJSON(http.StatusOK, response)
	}
	return
}
// swagger:operation DELETE /tenants/{tenant} Tenants DeleteTenants
// DELETE Tenants on the kubernetes
// ---
// produces:
// - application/json
// parameters:
//   - name: tenant
//     in: path
//     description: tenant looking for
//     required: true
//     type: string
// security:
//   - Bearer: []
// responses:
//     '200':
//         description: ok
//     '400':
//         description: Bad request
//     '500':
//         description: Internal server error
func RemoveTenant(c *gin.Context) {
	name := c.Param("name")
	if name == "admin" {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error":"Can't remove tenant admin"})
		return
	}
	name = "ogree-"+name
	if err := k8s.DeleteNamespace(name);err != nil{
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}else{
		c.IndentedJSON(http.StatusOK, gin.H{"success":"Successfully removed tenant "+name})
	}
}
// swagger:operation POST /tenants Tenants PostTenants
// Get Tenants on the kubernetes
// ---
// produces:
// - application/json
// parameters:
//   - name: body
//     in: body
//     description: 'Mandatory: name,customerPassword,hasWeb.'
//     required: true
//     format: object
//     example: '{"name": "super-tenants","customerPassword":"admin","hasWeb": true}'
// security:
//   - Bearer: []
// responses:
//     '200':
//         description: ok
//     '400':
//         description: Bad request
//     '500':
//         description: Internal server error
func AddTenant(c *gin.Context){
	var data models.Tenant
	// Call BindJSON to bind the received JSON to
	if err := c.ShouldBindJSON(&data); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ns := "ogree-"+data.Name
	if err := k8s.CreateNamespace(ns); err != nil{
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	//Mongo-db
	mongodb := newDeployment()
	mongodb.Image.Tag="latest"
	mongodb.Image.Repository="registry.ogree.ditrit.io/mongo-db-core"
	mongodb.Ingress.Enabled=false
	mongodb.Service.Port=27017
	if err := SetDeployement("mongo-db",data,mongodb,false);err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	//Mongo-API
	mongo := newDeployment()
	mongo.Image.Tag=data.ImageTag
	mongo.Image.Repository="registry.ogree.ditrit.io/mongo-api"
	mongo.Ingress.Enabled=!data.HasBFF
	mongo.Service.Port=3001
	mongo.Ingress.Hosts[0].Host="api"
	if err := SetDeployement("mongo-api",data,mongo,false);err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	//Web
	if data.HasWeb {
		web := newDeployment()
		web.Image.Tag=data.ImageTag
		web.Image.Repository="registry.ogree.ditrit.io/ogree-app"
		web.Ingress.Enabled=true
		web.Service.Port=80
		web.Ingress.Hosts[0].Host="app"
		if err := SetDeployement("app",data,web,false);err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		if err:= setLogo(data.Name);err!=nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

	}

	if data.HasBFF{
		// Create arango db
		arangodb := newDeployment()
		arangodb.Image.Tag="3.11.2"
		arangodb.Image.Repository="arangodb/arangodb"
		arangodb.Ingress.Enabled=false
		arangodb.Service.Port=8529
		if err := SetDeployement("arango-db",data,arangodb,false);err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		// Create Arango API
		arango := newDeployment()
		arango.Image.Tag=data.ImageTag
		arango.Image.Repository="registry.ogree.ditrit.io/arango-api"
		arango.Ingress.Enabled=false
		arango.Service.Port=8080
		if err := SetDeployement("arango-api",data,arango,false);err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		//Create BFF
		bff := newDeployment()
		bff.Image.Tag=data.ImageTag
		bff.Image.Repository="registry.ogree.ditrit.io/ogree-bff"
		bff.Ingress.Enabled=true
		bff.Ingress.Hosts[0].Host="api"
		bff.Service.Port=8085
		if err := SetDeployement("ogree-bff",data,bff,false);err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
		}
	}

	

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Tenant succefully created"})
}

func setLogo(name string)error {
	if _, err := os.Stat(name+".png");err == nil {
		if podName,err := k8s.GetPodName("app","ogree-"+name); err != nil{
			return err
		}else{
			command:="kubectl cp "+name+".png ogree-"+name+"/"+podName+":/usr/share/nginx/html/assets/assets/custom/logo.png"
			for {
				phase,err := k8s.GetPodPhase("app","ogree-"+name)
				if err != nil {
					 return err
				}
				if strings.Contains(phase,"run"){
					break
				}
				time.Sleep(1*time.Second)
			}
			if _, err := cmd.ExecCommand(command); err != nil {
				return err
			}
			command="rm -rf "+name+".png"
			cmd := exec.Command("bash","-c", command)
			if err := cmd.Run(); err != nil {
				return err
			}
			
		}
		
	}
	return nil
}
// swagger:operation PUT /tenants/{tenant} Tenants PutTenants
// Get Tenants on the kubernetes
// ---
// produces:
// - application/json
// parameters:
//   - name: tenant
//     in: path
//     description: tenant looking for
//     required: true
//     type: string
//   - name: body
//     in: body
//     description: 'Mandatory: name,customerPassword,hasWeb.'
//     required: true
//     format: object
//     example: '{"name": "super-tenants","customerPassword":"admin","hasWeb": true}'
// security:
//   - Bearer: []
// responses:
//     '200':
//         description: ok
//     '400':
//         description: Bad request
//     '500':
//         description: Internal server error
func UpdateTenants(c *gin.Context){
	var data models.Tenant
	// Call BindJSON to bind the received JSON to
	if err := c.ShouldBindJSON(&data); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	name := c.Param("name")
	if name == "admin" {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error":"Can't update tenant admin"})
		return
	}
	if tenant,err := k8s.GetTenantStatus(name); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}else{
		data.CustomerPassword = tenant.CustomerPassword
		//Mongo-API
		mongo := newDeployment()
		mongo.Image.Tag=data.ImageTag
		mongo.Image.Repository="registry.ogree.ditrit.io/mongo-api"
		mongo.Ingress.Enabled=!data.HasBFF
		mongo.Service.Port=3001
		mongo.Ingress.Hosts[0].Host="api"
		if err := SetDeployement("mongo-api",data,mongo,true);err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		//Web
		if data.HasWeb {
			web := newDeployment()
			web.Image.Tag=data.ImageTag
			web.Image.Repository="registry.ogree.ditrit.io/ogree-app"
			web.Ingress.Enabled=true
			web.Service.Port=80
			web.Ingress.Hosts[0].Host="app"
			if err := SetDeployement("app",data,web,tenant.HasWeb);err != nil {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
		}else if tenant.HasWeb{
			if err := helm.Uninstall("app","ogree-"+name);err != nil {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
		}

		if data.HasBFF{
			// Create arango db
			arangodb := newDeployment()
			arangodb.Image.Tag="3.11.2"
			arangodb.Image.Repository="arangodb/arangodb"
			arangodb.Ingress.Enabled=false
			arangodb.Service.Port=8529
			if err := SetDeployement("arango-db",data,arangodb,tenant.HasBFF);err != nil {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}

			// Create Arango API
			arango := newDeployment()
			arango.Image.Tag=data.ImageTag
			arango.Image.Repository="registry.ogree.ditrit.io/arango-api"
			arango.Ingress.Enabled=false
			arango.Service.Port=8080
			if err := SetDeployement("arango-api",data,arango,tenant.HasBFF);err != nil {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
			//Create BFF
			bff := newDeployment()
			bff.Image.Tag=data.ImageTag
			bff.Image.Repository="registry.ogree.ditrit.io/ogree-bff"
			bff.Ingress.Enabled=true
			bff.Ingress.Hosts[0].Host="api"
			bff.Service.Port=8085
			if err := SetDeployement("ogree-bff",data,bff,tenant.HasBFF);err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
			}

		}/*else if tenant.HasBFF {
			if err := helm.Uninstall("arango-db","ogree-"+name);err != nil {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
			if err := helm.Uninstall("arango-api","ogree-"+name);err != nil {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
			if err := helm.Uninstall("ogree-bff","ogree-"+name);err != nil {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
		}*/
	}


	c.IndentedJSON(http.StatusOK, gin.H{"message": "Tenant succefully updated"})
}


func BackupTenantDB(c *gin.Context) {
	name := c.Param("name")
	if name == "admin" {
		c.IndentedJSON(http.StatusForbidden, gin.H{"error":"No database in admin"})
		return
	}
	// Call BindJSON to bind the received JSON
	var backupInfo models.Backup
	if err := c.BindJSON(&backupInfo); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	if filename, err :=k8s.CreateMongoArchive("ogree-"+name,backupInfo.DBPassword);err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}else{
		c.File(filename)
		command := "rm -rf "+filename
		cmd.ExecCommand(command)
	}
}


func AddTenantLogo(c *gin.Context){
	name := c.Param("name")
	formFile, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	err = c.SaveUploadedFile(formFile, name+".png")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	c.String(http.StatusOK, "")

}
func newDeployment() (models.Deployement){
	return models.Deployement{
		ReplicaCount: 1,
		Name: "",
		ImagePullSecrets: []models.ImagePullSecrets{{Name:"regcred",}},
		Image: models.Image{Tag: "",
							Repository:"",
							PullPolicy: "Always",},
		Ingress: models.Ingress{Enabled: false,
								EntryPoints: []string{"web"},
								Hosts: []models.Host{{Host: "host",}},},
		SecurityContext: nil,
	}
}

func SetDeployement(name string,data models.Tenant, deploy models.Deployement,install bool) error{
	ns := "ogree-"+data.Name
	var env []models.Env
	var configmap []models.ConfigMap
	var volumes []models.PersistentVolumeClaim
	if err := helm.GetDataFromYaml("env/"+name+".yaml",&env);err != nil{
		return err
	}else{

		if err := helm.GetDataFromYaml("configmaps/"+name+".yaml",&configmap);err != nil{
			return err
		}else{
			if err := helm.GetDataFromYaml("volume/"+name+".yaml",&volumes);err != nil{
				return err
			}else{
				deploy.ConfigMap = configmap
				deploy.Env = env
				deploy.PersistentVolumeClaim = volumes
				deploy.Name = name
				if err:= helm.WriteDeployementFile(deploy,data);err != nil{
					return err
				}
				if err := helm.InstallYaml(name,ns,install);err != nil{
					return err
				}
			}
		}
	}
	return nil
}