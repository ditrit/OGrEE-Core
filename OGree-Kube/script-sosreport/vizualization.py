import json
import sys
import re
import subprocess


def get_cluster_name():
    try:
        # Run kubectl command to get the cluster name
        result = subprocess.run(["kubectl", "config", "current-context"], capture_output=True, text=True, check=True)
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        print("Error:", e)
        return None


def get_pod_pvs(namespace, pod_name):
    try:
        # Run the kubectl command to get the pods in the specified namespace
        result = subprocess.run(["kubectl", "get", "pods", "-n", namespace], capture_output=True, text=True, check=True)
        
        # Filter the output using grep based on the pod name pattern
        filtered_pods = subprocess.run(["grep", f'^{pod_name}.*'], input=result.stdout, capture_output=True, text=True, check=True)
        
        # Split the output into lines and get the first line (assuming there's only one matching pod)
        matched_pod_line = filtered_pods.stdout.strip().split('\n')[0]
        
        # Extract the pod name from the matched line
        matched_pod_name = matched_pod_line.split()[0]
        
        # Run kubectl command to get the pod details in JSON format
        result = subprocess.run(["kubectl", "-n", namespace, "get", "pod", matched_pod_name, "-o", "json"], capture_output=True, text=True, check=True)
        pod_info = json.loads(result.stdout)
        
        volumes = pod_info["spec"].get("volumes", [])
        pvs = []
        for volume in volumes:
            if volume.get("persistentVolumeClaim"):
                pvc_name = volume["persistentVolumeClaim"]["claimName"]
                pv_info = subprocess.run(["kubectl", "-n", namespace, "get", "pvc", pvc_name, "-o", "json"], capture_output=True, text=True, check=True)
                pvc_info = json.loads(pv_info.stdout)
                pv_name = pvc_info["spec"]["volumeName"]
                pv_details = subprocess.run(["kubectl", "get", "pv", pv_name, "-o", "json"], capture_output=True, text=True, check=True)
                pv = json.loads(pv_details.stdout)
                pv_type = pv.get("spec", {}).get("storageClassName", "Unknown")
                pv_path = ""
                pv_ip = ""
                if pv_type == "nfs":
                    pv_path = pv["spec"]["nfs"]["path"]
                    pv_ip = pv["spec"]["nfs"]["server"]
                elif pv_type == "azureFile":
                    pv_path = pv["spec"]["azureFile"]["secretName"]
                elif pv_type == "awsElasticBlockStore":
                    pv_path = pv["spec"]["awsElasticBlockStore"]["volumeID"]
                elif pv_type == "local-path":
                    pv_path = pv["spec"]["hostPath"]["path"]
                    pv_ip = "None"
                elif pv_type == "csi-s3":
                    pv_path = pv["spec"]["csi"]["volumeHandle"]
                    pv_ip = "None"                
                else:
                    pv_type = "Unknown"
                    pv_path = "None"
                    pv_ip = "None"
                pv_info = {
                    "pv-type": pv_type,
                    "pv-path": pv_path,
                    "pv-ip": pv_ip
                }
                pvc_info = {
                    "pvc": pvc_name,
                    "children": [pv_info]
                }
                pvs.append(pvc_info)
        return pvs
    except subprocess.CalledProcessError as e:
        print("Error:", e)
        return None


def format_json(input_file, output_file):
    with open(input_file, 'r') as f:
        data = json.load(f)

    node_names = [node["metadata"]["name"] for node in data.get("items", []) if node["kind"] == "Node"]

    formatted_data = {
        "nodes": node_names,
        "namecluster": get_cluster_name(),
        "children": []
    }

    for item in data.get("items", []):
        if item["kind"] == "Deployment":
            deployment_info = {
                "namespace": item["metadata"]["namespace"],
                "type": "deployment",
                "name": item["metadata"]["name"],
                "children": []
            }
            for pod_status in item.get("status", {}).get("conditions", []):
                if pod_status["type"] == "Progressing":
                    message = pod_status.get("message", "")
                    match = re.search(r'ReplicaSet "([^"]+)" has successfully progressed.', message)
                    if match:
                        pod_name = match.group(1)
                        pod_info = {
                            "type": "pod",
                            "name": pod_name,
                            "pv_name": get_pod_pvs(item["metadata"]["namespace"], pod_name)
                        }
                        deployment_info["children"].append(pod_info)
            formatted_data["children"].append(deployment_info)

    with open(output_file, 'w') as f:
        json.dump(formatted_data, f, indent=4)


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python script.py <input_file.json> <output_file.json>")
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2]
    format_json(input_file, output_file)
