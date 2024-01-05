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
      pods=$(kubectl get pods -n "$namespace" -l k8s-app="$service" -o jsonpath='{range .items[*]}{.metadata.name}{" "}{end}' | tr ' ' '\n')

      for pod in $pods; do
        # Get PV information (currently empty, modify as needed)
        pv_location=$(kubectl get pod "$pod" -n "$namespace" -o jsonpath='{.spec.volumes[0].persistentVolumeClaim.volumeName}' 2>/dev/null)
        if [ -z "$pv_location" ]; then
          pv_location="local"
        else
          pv_location="external"
        fi

        namespace_json+="{ 
          \"category\": \"application\",
          \"type\": \"pod\",
          \"name\": \"$pod\",
          \"pv_location\": \"$pv_location\"
        },"
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