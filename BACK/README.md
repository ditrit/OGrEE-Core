# Backend
BACK contains two backend applications for deploying and managing tenants. The frontend APP uses these backends to enter "SuperAdmin" mode, it can communicate seamlessly with either of them. 

Docker Backend uses docker compose to launch a tenant while Kube Backend uses kubernetes, creating a new namespace for each tenant. 