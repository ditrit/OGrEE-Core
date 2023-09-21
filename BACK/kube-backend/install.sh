#!/bin/bash
# USAGE : ./install.sh {profile} {DNS} {PORT_DNS}
if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run with sudo or as root."
    exit 1
fi

if [ "$#" -ne 2 ];then 
    echo "USAGE : sudo $0 {DNS} {PORT_DNS}"
    exit 1
fi

## Install helm 
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh


## Install traefik-v2
kubectl apply -f traefik/traefik-CRDs.yml 
kubectl create ns traefik-v2
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm install -n traefik-v2 traefik traefik/traefik --set service.type=NodePort

helm install -n traefik-v2 dashboard ./traefik/dashboard --set host=$1


## Get redicrect url for reverse proxy
KUBE_IP=$(ip addr show eth0 | grep -oP '(?<=inet\s)\d+\.\d+\.\d+\.\d+')
TRAEFIK_PORT=$(kubectl get service -n traefik-v2 traefik -o jsonpath={.spec.ports[0].nodePort})
REDIRECT_URL=http://$KUBE_IP:$TRAEFIK_PORT

## init nginx reverse proxy
apt update -y
apt install nginx -y
NGINX_CONF=$(cat << EOF  
server {
        listen $2;
        server_name *.$1;
        client_max_body_size 250M;
        location / {
                proxy_set_header Host            \$host;
                proxy_set_header X-Forwarded-For \$remote_addr;
                proxy_pass $REDIRECT_URL;
  }
}
EOF
)
apt-get install ed -y 
ed -s /etc/nginx/nginx.conf << EOF
/^http {/
/^http {/a
$NGINX_CONF
.
w
q
EOF

systemctl restart nginx
systemctl enable nginx



## install admin API
kubectl create ns ogree-admin

kubectl create secret generic regcred \
        --from-file=.dockerconfigjson=kube-admin/helm/docker/config.json \
        --type=kubernetes.io/dockerconfigjson \
        -n ogree-admin

kubectl apply -f svc/sa.yaml -n ogree-admin

helm install admin kube-admin/helm/ogree/ \
    -f kube-admin/helm/admin-values.yaml \
    --set env[0].name=HOST \
    --set env[0].value=$1 \
    --set ingress.hosts[0].host=api.admin.$1 \
    -n ogree-admin


## Install ogree-app admin
helm install app kube-admin/helm/ogree \
    -f kube-admin/helm/app-admin-values.yaml \
    --set ingress.hosts[0].host=app.admin.$ \
    --set configmap[0].data[0].value=API_URL=https://api.admin.$1$'\n'ALLOW_SET_BACK=true$'\n'BACK_URLS=https://api.admin.$1 \
    -n ogree-admin




