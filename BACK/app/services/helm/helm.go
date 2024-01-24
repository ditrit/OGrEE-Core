package helm

import (
	"back-admin/models"
	"back-admin/services/cmd"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func GetDataFromYaml(filename string, output interface{}) error {

	file, err := os.Open("helm/" + filename) // For read access.
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return nil
		}
		return err
	}
	data := make([]byte, 2048)
	count, err := file.Read(data)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data[:count], output)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func DeploytoYaml(filename string, deploy models.Deployement) error {

	b, err := yaml.Marshal(&deploy)
	if err != nil {
		return err
	}

	err = os.WriteFile("helm/"+filename, b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func WriteDeployementFile(deploy models.Deployement, data models.Tenant) error {
	if deploy.Ingress.Enabled {
		deploy.Service.Type = "NodePort"
		deploy.Ingress.Hosts[0].Host += "." + data.Name + "." + os.Getenv("HOST")
	} else {
		deploy.Service.Type = "ClusterIP"
	}
	setBDDPassword(deploy.Env, data.CustomerPassword)
	setAPPUrl(deploy.ConfigMap, data.Name)
	if err := DeploytoYaml("values.yaml", deploy); err != nil {
		return err
	}
	return nil
}

func setAPPUrl(config []models.ConfigMap, name string) {
	for i := range config {
		if config[i].Name == "webapp-env" {
			h := "https://"
			if os.Getenv("HOST") == "localhost" {
				h = "http://"
			}
			apiUrl := h + "api." + name + "." + os.Getenv("HOST")
			configValue := "API_URL=" + apiUrl + "\nALLOW_SET_BACK=false"
			config[i].Data[0].Value = configValue
		}
	}
}

func setBDDPassword(env []models.Env, pass string) {
	for i := range env {
		if env[i].Name == "db_pass" ||
			env[i].Name == "ARANGO_PASSWORD" ||
			env[i].Name == "CUSTOMER_API_PASSWORD" ||
			env[i].Name == "ARANGO_ROOT_PASSWORD" {
			env[i].Value = pass
		}
	}
}

func InstallYaml(name, ns string, install bool) error {
	verb := "install"
	if install {
		verb = "upgrade"
	}
	command := "helm " + verb + " " + name + " helm/ogree -f " + "helm/values.yaml -n " + ns
	_, err := cmd.ExecCommand(command)
	if err != nil {
		return err
	}
	return nil
}

func Uninstall(name, ns string) error {
	command := "helm uninstall " + name + " -n " + ns
	_, err := cmd.ExecCommand(command)
	if err != nil {
		return err
	}
	return nil
}
