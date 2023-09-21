# Install a Kubernetes cluster
## On All Nodes

### Port
You may needs to open some ports like 
- 6443/tcp
- 2379/tcp
- 2380/tcp
- 10250/tcp
- 10251/tcp
- 10252/tcp
- 10255/tcp
- 30000:32767/tcp

### Swap
Kubernetes needs to shutdown swap to work properly

```bash
sudo swapoff -a
sudo sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
```

### ContainerD

Before installing containerd, set the following kernel parameters on all the nodes.

```bash
$ cat <<EOF | sudo tee /etc/modules-load.d/containerd.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter

cat <<EOF | sudo tee /etc/sysctl.d/99-kubernetes-k8s.conf
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF

sudo sysctl --system
```
Now, install conatinerd by running following apt command on all the nodes.

```bash
sudo apt  update
sudo apt -y install containerd

containerd config default | sudo tee /etc/containerd/config.toml >/dev/null 2>&1
```
Set cgroupdriver to systemd on all the nodes,

Edit the file ‘/etc/containerd/config.toml’ and look for the section ‘[plugins.”io.containerd.grpc.v1.cri”.containerd.runtimes.runc.options]’ and add SystemdCgroup = true

If you want to use another default partition for our cluster, you can set section root and state

Restart and enable containerd service on all the nodes,
```bash
sudo systemctl restart containerd
sudo systemctl enable containerd
```

### Kubernetes

```bash
sudo apt install gnupg gnupg2 curl software-properties-common -y
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmour -o /etc/apt/trusted.gpg.d/cgoogle.gpg
sudo apt-add-repository "deb http://apt.kubernetes.io/ kubernetes-xenial main"

sudo apt update
sudo apt install kubelet kubeadm kubectl -y
sudo apt-mark hold kubelet kubeadm kubectl
```

## On Master

Now, we are all set to create Kubernetes cluster, run following command only from master node,

```bash
sudo kubeadm init
```
To start interacting with cluster, run following commands on master node,
```bash
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```
Run following kubectl command to get nodes and cluster information,
```bash
kubectl get nodes
kubectl cluster-info
```

### Storage-Class

You need to create a local Storage class if you want to use volumes:

```bash
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

## On Worker

kubeadm command on master will give you a command line on output like
```bash
kubeadm join k8s-master:6443 --token <token> \
	--discovery-token-ca-cert-hash sha256:<sha256> 
```
Exectute this command line on all worker to join master

# Kube-Admin

Like Ogree SuperAdmin, Kube-Admin is able to manage ogree tenants on a kubernetes cluster

It's will use a custer service Account to interracte with the cluster:
- SVC file in on svc/sa.yaml

A new tenants corresponding to a kubernetes namespace, so you have to be able to manage it

## helm
Kube-admin use helm and template to manager deployement, it's will use kube-admin/helm/ogree template for all deployement

## Traefik

As ingress, kube-admin use traefik, you need to configure traefik into kubernetes cluster

```bash
kubectl create ns traefik-v2
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm install -n traefik-v2 traefik traefik/traefik --set service.type=NodePort
```
## Install

To install ogree kube-admin on cluster, you can use `install.sh` script, it's will:
- Install helm 
- Install traefik on cluster
- Add traefik to a nginx reverse-proxy on the machine
- Create ogree-admin namespace
- Install ogree service account and role binding
- Install kube-admin and ogree-app

Usage:
 - DNS: CNAME of the cluster
 - PORT_DNS: port listen by nginx reverse-proxy

```bash
sudo install.sh {DNS} {PORT_DNS}
```

## Update a pod on kubernetes

```bash
kubectl set image deploy/<deploy_name> <pode_name>=<registry>/<image>:<tag>
```

Example:
```bash
kubectl set image deploy/kube-admin kube-admin=registry.ogree.ditrit.io/kube-admin:0.4.3
```
