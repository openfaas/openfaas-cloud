
# Testing the OpenFaaS Cloud Builder with Kubernetes

This tutorial allows you to test the OpenFaaS Cloud Builder (of-builder) without having OpenFaaS Cloud installed or deployed.

## Create a namespace

```sh
kubectl create ns openfaas
```

## Log into `docker`

Make sure keychain access is disabled, then force a regeneration of your `config.json` file

```sh
export DOCKER_USERNAME="your-username"

rm $HOME/.docker/config.json
docker login $DOCKER_USERNAME

# Check the file is base64 encoded

cat $HOME/.docker/config.json
```

> Note: If you are having issues with your config.json file, then try [these instructions](https://github.com/openfaas-incubator/ofc-bootstrap#prepare-your-docker-registry).

## Create a registry secret

```sh
# Delete if you already have one
kubectl delete secret -n openfaas registry-secret

export SERVER="https://index.docker.io/v1/"
export DOCKER_USERNAME="your-username"
export DOCKER_PASSWORD="your-pass"

kubectl create secret generic registry-secret \
 -n openfaas \
 --from-file $HOME/.docker/config.json
```

## Create a dummy payload secret

We'll disable the payload validation by creating a dummy payload secret:

```sh
kubectl create secret generic payload-secret \
  -n openfaas \
  --from-literal payload-secret=""
```

## Clone the YAML files and edit

```sh
git clone https://github.com/openfaas/openfaas-cloud
cd openfaas-cloud/yaml/core/
```

Open `of-builder-dep.yml` in a text editor.

* Edit `disable_hmac`, set the value to `true`

  This prevents the need to sign the payload, do not run with this configuration outside of dev/test.

* Edit `enable_lchown`, set the value to `false`

## Deploy

```sh
kubectl apply -f yaml/core/of-builder-svc.yml
kubectl apply -f yaml/core/of-builder-dep.yml

# Wait to come up

kubectl rollout status -n openfaas deployment.apps/of-builder
```

## Create a new function

```sh
# Work in a temporary directory
mkdir -p /tmp/builder
cd /tmp/builder

export FN="go-tester"

faas-cli new --lang go $FN
faas-cli build --shrinkwrap -f $FN.yml
```

Edit it if you wish.

## Create your build context

Build the function:

```sh
export DOCKER_USERNAME="your-username"
export SERVER="docker.io/"

rm -fr tmp
mkdir -p tmp/context

echo '{"Ref": "'${SERVER}${DOCKER_USERNAME}'/'${FN}':latest"}' > tmp/com.openfaas.docker.config

cp -r build/$FN/* tmp/context
tar -C ./tmp -cvf $FN-context.tar .
```

## Run the build

```sh
kubectl port-forward -n openfaas svc/of-builder 8081:8080 &

time curl -i --data-binary @$FN-context.tar http://127.0.0.1:8081/build

# View as JSON
curl -s --data-binary @$FN-context.tar http://127.0.0.1:8081/build | jq
```

## Check the build on the Docker Hub

```sh
echo Find your image at: https://hub.docker.com/r/$DOCKER_USERNAME/$FN
```

## Deploy the function

If you have OpenFaaS deployed you can now deploy the function

```sh
kubectl port-forward -n openfaas svc/gateway 31112:8080
export OPENFAAS_URL=127.0.0.1:31112

PASSWORD=$(kubectl get secret -n openfaas basic-auth -o jsonpath="{.data.basic-auth-password}" | base64 --decode; echo)
echo -n $PASSWORD | faas-cli login --username admin --password-stdin

faas-cli deploy --name $FN --image "${DOCKER_USERNAME}/${FN}:latest"

echo | faas-cli invoke $FN
```
