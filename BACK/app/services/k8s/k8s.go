package k8s

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"kube-admin/models"
	"kube-admin/services/cmd"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func GetNamespace() ([]string, error) {

	command := "kubectl get namespace -o jsonpath={.items}"
	ns, err := cmd.ExecCommand(command)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	var namespaces []models.Namespace
	err = json.Unmarshal([]byte(ns), &namespaces)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	var result []string
	for _, ns := range namespaces {
		if strings.Contains(ns.Metadata.Name, "ogree-") {
			result = append(result, strings.TrimPrefix(ns.Metadata.Name, "ogree-"))
		}
	}
	return result, nil

}
func NetboxCreated() bool {
	command := "kubectl get ns netbox-ogree -o jsonpath={.metadata.name}"
	result, err := cmd.ExecCommand(command)
	if err != nil {
		return false
	}
	if strings.Contains(result, "netbox") {
		return true
	}
	return false
}

func GetNetbox() (models.ContainerInfo, error) {
	name, err := GetPodName("netbox", "netbox-ogree")
	if err != nil {
		return models.ContainerInfo{}, err
	}
	command := "kubectl get pods " + name + " -n netbox-ogree -o json"
	pods, err := cmd.ExecCommand(command)
	if err != nil {
		return models.ContainerInfo{}, errors.New(err.Error())
	}
	var po models.Pod
	err = json.Unmarshal([]byte(pods), &po)
	if err != nil {
		fmt.Println(err.Error())
		return models.ContainerInfo{}, errors.New(err.Error())
	}
	pod := models.ContainerInfo{
		Name:        po.Spec.Containers[0].Name,
		Status:      strings.ToLower(po.Status.Phase),
		Image:       po.Spec.Containers[0].Image,
		Ports:       strconv.Itoa(po.Spec.Containers[0].Ports[0].ContainerPort),
		Size:        "unknown",
		LastStarted: po.Status.StartTime,
	}

	return pod, nil

}

func CreateNamespace(name string) error {
	//Create namespace
	command := "kubectl create namespace " + name
	_, err := cmd.ExecCommand(command)
	if err != nil {
		return err
	}
	// add docker cretentials to namespace
	dockerCred := "kubectl create secret generic regcred --from-file=.dockerconfigjson=./helm/docker/config.json --type=kubernetes.io/dockerconfigjson -n " + name
	_, err = cmd.ExecCommand(dockerCred)
	if err != nil {
		return err
	}
	return nil
}

func DeleteNamespace(name string) error {
	command := "kubectl delete namespace " + name
	_, err := cmd.ExecCommand(command)
	if err != nil {
		return err
	}
	return nil
}

func GetPods(ns string) ([]models.ContainerInfo, error) {
	command := "kubectl get pod -n " + ns + " -o jsonpath={.items}"
	pods, err := cmd.ExecCommand(command)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	var po []models.Pod
	err = json.Unmarshal([]byte(pods), &po)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New(err.Error())
	}

	var podsInfo []models.ContainerInfo

	for _, p := range po {
		pod := models.ContainerInfo{
			Name:        ns + "_" + p.Spec.Containers[0].Name,
			Status:      strings.ToLower(p.Status.Phase),
			Image:       p.Spec.Containers[0].Image,
			Ports:       strconv.Itoa(p.Spec.Containers[0].Ports[0].ContainerPort),
			Size:        "unknown",
			LastStarted: p.Status.StartTime,
		}
		podsInfo = append(podsInfo, pod)
	}
	return podsInfo, nil
}
func getDeployEnv(name, ns string) ([]models.Env, error) {
	command := "kubectl get deploy " + name + " -n " + ns + " -o jsonpath={.spec.template.spec.containers[].env}"
	var result []models.Env
	env, err := cmd.ExecCommand(command)
	if err != nil {
		return result, errors.New(err.Error())
	}
	err = json.Unmarshal([]byte(env), &result)
	if err != nil {
		return result, errors.New(err.Error())
	}
	return result, nil
}

func GetTenantStatus(name string) (models.Tenant, error) {
	var tenant models.Tenant
	tenant.Name = name
	if pods, err := GetPods("ogree-" + name); err != nil {
		return tenant, err
	} else {
		for _, pod := range pods {
			if pod.Name == "app" {
				tenant.HasWeb = true
			}
			if pod.Name == "ogree-bff" {
				tenant.HasBFF = true
			}
			if pod.Name == "mongo-api" {
				if envs, err := getDeployEnv(pod.Name, "ogree-"+tenant.Name); err != nil {
					return tenant, err
				} else {
					for _, env := range envs {
						if env.Name == "db_pass" {
							tenant.CustomerPassword = env.Value
						}
					}
				}

			}
		}
	}
	return tenant, nil
}

func GetPodName(deploy, ns string) (string, error) {

	command := "kubectl get pods -o jsonpath='{.items[*].metadata.name}' -n " + ns
	podsName, err := cmd.ExecCommand(command)
	if err != nil {

		return "", errors.New(err.Error())
	}
	names := strings.Split(podsName, " ")
	for _, name := range names {
		if strings.Contains(name, deploy) {

			name = strings.Replace(name, "'", "", -1)
			return name, nil
		}
	}
	return "", errors.New("No pods found for " + deploy)
}

func GetPodPhase(deploy, ns string) (string, error) {
	if name, err := GetPodName(deploy, ns); err != nil {
		return "", errors.New(err.Error())
	} else {
		command := "kubectl get pods " + name + " -o jsonpath='{.status.containerStatuses[].state}' -n " + ns
		if phase, err := cmd.ExecCommand(command); err != nil {
			return "", errors.New(err.Error())
		} else {
			return phase, nil
		}
	}
}

func CreateMongoArchive(ns, password string) (string, error) {
	t := time.Now()
	filename := ns + "_db_" + t.Format("2006-01-02T150405") + ".archive"
	if name, err := GetPodName("mongo-db", ns); err != nil {
		return "", err
	} else {
		command := "kubectl exec -n " + ns + " " + name + " -- mongodump --username  ogreeogree-coreAdmin --password " + password + " -d ogreeogree-core --archive > " + filename
		cmd := exec.Command("bash", "-c", command)
		fmt.Println(cmd.Args)

		var stderr bytes.Buffer
		cmd.Dir = "."
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return "", err
		}
	}
	return filename, nil
}

func GetContainerLogs(ns, name string) (string, error) {
	pod, err := GetPodName(name, ns)
	if err != nil {
		return "", err
	}
	command := "kubectl logs -n " + ns + " pods/" + pod
	logs, err := cmd.ExecCommand(command)
	if err != nil {
		return "", errors.New(err.Error())
	}
	return logs, nil
}
