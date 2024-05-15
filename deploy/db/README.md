# MongoDB as a replicaSet

The scripts defined here are used to start a MongoDB instance configured as a replicaSet. The Dockerfile defines the docker image of a mongo replicaSet that will be used in Kubernetes (necessary for the Kubernetes version of BACK).

To build the docker image we need to run the `docker build` command at the `deploy` directory.

```
# In the deploy directory, build the Docker image
docker build -f db/Dockerfile -t <tag_name> .
```

**WARNING:** do not build the image using a machine with ARM architecture (for example all Macs with chip M*) as it will not work in our cluster. 