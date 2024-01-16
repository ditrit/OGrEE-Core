#!/bin/bash

# Function to get the current cluster name
get_cluster_name() {
  CLUSTER_NAME=$(kubectl config current-context)
}

# Function to check if a namespace exists
namespace_exists() {
  local namespace="$1"
  kubectl get namespace "$namespace" --ignore-not-found=true &>/dev/null
}

# Function to check if a namespace has services
has_services() {
  local namespace="$1"
  local services_count=$(kubectl get services -n "$namespace" --ignore-not-found=true --no-headers=true 2>/dev/null | wc -l)
  [ "$services_count" -gt 0 ]
}

# Function to generate JSON for a namespace and its resources
generate_namespace_json() {
  local namespace="$1"

  # Check if the namespace exists and has services
  if namespace_exists "$namespace" && has_services "$namespace"; then
    namespace_json="
       "

    # Get services in the namespace
    services=$(kubectl get services -n "$namespace" -o jsonpath='{range .items[*]}{.metadata.name}{" "}{end}' | tr ' ' '\n')

    for service in $services; do
      namespace_json+="{ 
        \"namespace_kube\": \"$namespace\",
        \"category\": \"application\",
        \"type\": \"kube-application\",
        \"name\": \"$service\",
        \"children\": ["

      # Get pods in the service
      pods=$(kubectl get pods -n "$namespace" -l app.kubernetes.io/name="$service" -o jsonpath='{range .items[*]}{.metadata.name}{" "}{end}' | tr ' ' '\n')

      for pod in $pods; do
        namespace_json+="{ 
          \"category\": \"application\",
          \"type\": \"pod\",
          \"name\": \"$pod\",
          \"children\": ["

        # Get a list of PVCs for the current pod
        pvc_names=$(kubectl get pods "$pod" -n "$namespace" -o=jsonpath='{.spec.volumes[*].persistentVolumeClaim.claimName}')

        # Loop through each PVC
        for pvc in $pvc_names; do
          namespace_json+="{ 
            \"pvc\": \"$pvc\",
            \"children\": ["

          # Get PV information for the current PVC
          pv_info=$(kubectl get pvc "$pvc" -n "$namespace" -o=jsonpath='{.spec.volumeName},{.spec.storageClassName}')

          # Extract PV name and storage class
          pv_name=$(echo "$pv_info" | cut -d',' -f1)
          storage_class=$(echo "$pv_info" | cut -d',' -f2)

          # Get PV details based on storage class
          case $storage_class in
            "azure-disk")
              pv_details=$(kubectl get pv "$pv_name" -o=jsonpath='{.spec.volumeMode},{.spec.azureDisk.diskURI}')
              pv_type="AzureDisk"
              pv_path=$(echo "$pv_details" | cut -d',' -f2)
              pv_ip=""
              ;;
            "awsElasticBlockStore")
              pv_details=$(kubectl get pv "$pv_name" -o=jsonpath='{.spec.volumeMode},{.spec.awsElasticBlockStore.volumeID}')
              pv_type="ElasticBlockStore"
              pv_path=$(echo "$pv_details" | cut -d',' -f2)
              pv_ip=""
              ;;
            "CephFS")
              pv_details=$(kubectl get pv "$pv_name" -o=jsonpath='{.spec.volumeMode},{.spec.cephfs.monitors[0]},{.spec.cephfs.path}')
              pv_type="CephFS"
              pv_ip=$(echo "$pv_details" | cut -d',' -f2)
              pv_path=$(echo "$pv_details" | cut -d',' -f3)
              ;;
            "nfs-client")
              pv_details=$(kubectl get pv "$pv_name" -o=jsonpath='{.spec.volumeMode},{.spec.nfs.server},{.spec.nfs.path}')
              pv_type="NFS"
              pv_ip=$(echo "$pv_details" | cut -d',' -f2)
              pv_path=$(echo "$pv_details" | cut -d',' -f3)
              ;;
            "azureFile")
              pv_details=$(kubectl get pv "$pv_name" -o=jsonpath='{.spec.volumeMode},{.spec.azureFile.secretName},{.spec.azureFile.shareName}')
              pv_type="AzureFile"
              pv_path=$(echo "$pv_details" | cut -d',' -f2)
              pv_ip=""
              ;;
            "csi")
              pv_details=$(kubectl get pv "$pv_name" -o=jsonpath='{.spec.volumeMode},{.spec.csi.driver}')
              pv_type="CSI"
              pv_path=""
              pv_ip=""
              ;;
            "gcePersistentDisk")
              pv_details=$(kubectl get pv "$pv_name" -o=jsonpath='{.spec.volumeMode},{.spec.gcePersistentDisk.pdName}')
              pv_type="GCEPersistentDisk"
              pv_path=$(echo "$pv_details" | cut -d',' -f2)
              pv_ip=""
              ;;
            "glusterfs")
              pv_details=$(kubectl get pv "$pv_name" -o=jsonpath='{.spec.volumeMode},{.spec.glusterfs.endpoints-name},{.spec.glusterfs.path}')
              pv_type="GlusterFS"
              pv_ip=$(echo "$pv_details" | cut -d',' -f2)
              pv_path=$(echo "$pv_details" | cut -d',' -f3)
              ;;
            "fc")
              pv_details=$(kubectl get pv "$pv_name" -o=jsonpath='{.spec.volumeMode}')
              pv_type="FibreChannel"
              pv_path=""
              pv_ip=""
              ;;
            "cinder")
              pv_details=$(kubectl get pv "$pv_name" -o=jsonpath='{.spec.volumeMode},{.spec.cinder.volumeID}')
              pv_type="Cinder"
              pv_path=$(echo "$pv_details" | cut -d',' -f2)
              pv_ip=""
              ;;
            "iscsi")
              pv_details=$(kubectl get pv "$pv_name" -o=jsonpath='{.spec.volumeMode},{.spec.iscsi.targetPortal},{.spec.iscsi.iqn},{.spec.iscsi.lun}')
              pv_type="iSCSI"
              pv_ip=$(echo "$pv_details" | cut -d',' -f2)
              pv_path=$(echo "$pv_details" | cut -d',' -f3)
              ;;
            "local-path")
              pv_details=$(kubectl get pv "$pv_name" -o=jsonpath='{.spec.volumeMode},{.spec.hostPath.path}')
              pv_type="HostPath"
              pv_path=$(echo "$pv_details" | cut -d',' -f2)
              pv_ip="None"
              ;;
            *)
              pv_type="Unknown"
              pv_path="None"
              pv_ip="None"
              ;;
          esac

          # Print PV details
          namespace_json+="{ 
              \"pv-type\": \"$pv_type\",
              \"pv-path\": \"$pv_path\",
              \"pv-ip\": \"$pv_ip\"
            }"

          namespace_json+="]},"
        done

        # Remove comma
        namespace_json="${namespace_json%,}"

        namespace_json+="]},"
      done

      # Remove comma
      namespace_json="${namespace_json%,}"

      namespace_json+="] },"
    done

    # Remove comma
    namespace_json="${namespace_json%,}"
    #namespace_json+="}"
    echo "$namespace_json"
  else
    echo ""  # Return an empty string if the namespace does not exist or has no services
  fi
}

get_cluster_name

# Get the names of nodes dynamically
NODES=($(kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{" "}{end}' | tr ' ' '\n'))

# Generate JSON for the current cluster
cluster_json="
{ 
  \"namespace_cli\": \"logical\",
  \"device\": [\"${NODES[@]//\"/\"}\"],
  \"category\": \"application\", 
  \"type\": \"kubernetes\", 
  \"name\": \"$CLUSTER_NAME\", 
  \"children\": ["

# Get namespaces in the cluster
namespaces=$(kubectl get namespaces -o jsonpath='{range .items[*]}{.metadata.name}{" "}{end}' | tr ' ' '\n')

first_namespace=true
for namespace in $namespaces; do
  namespace_json="$(generate_namespace_json "$namespace")"
  if [ -n "$namespace_json" ]; then
    if [ "$first_namespace" = true ]; then
      first_namespace=false
    else
      cluster_json+=","
    fi
    cluster_json+="$namespace_json"
  fi
done

cluster_json+="]}"

# Print the final JSON for the cluster
echo "$cluster_json" | jq .