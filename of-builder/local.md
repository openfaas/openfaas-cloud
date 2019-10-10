
# Testing the OpenFaaS Cloud Builder with Kubernetes

This tutorial allows you to test the OpenFaaS Cloud Builder (of-builder) without having OpenFaaS Cloud installed or deployed.

## Log into `docker`  (Normal registry)

Make sure keychain access is disabled, then force a regeneration of your `config.json` file

```sh
export DOCKER_USERNAME="your-username"

rm $HOME/.docker/config.json
docker login $DOCKER_USERNAME

# Check the file is base64 encoded

cat $HOME/.docker/config.json
```

Now create the secret:

```sh
kubectl create secret generic registry-secret \
  -n openfaas \
 --from-file $HOME/.docker/config.json
```

## (Or) Create a registry secret for ECR

```sh
export DOCKER_USERNAME=""
export REGION="eu-central-1"
export ACCOUNT_ID="012345678900"

export SERVER="$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com/"

cat <<EOF > docker.config
{
  "credsStore": "ecr-login",
  "credHelpers": {
    "$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com": "ecr-login"
  }
}
EOF
```

Now create the secret:

```sh
kubectl create secret generic registry-secret \
  -n openfaas \
 --from-file config.json
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

* Edit `disable_hmac`, set the value to `false`
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
```

Edit it if you wish.

## Create your build context (for a normal registry)

Build the function:

```sh
export DOCKER_USERNAME=""
export SERVER="docker.io/"

faas-cli build --shrinkwrap -f $FN.yml

rm -fr tmp
mkdir -p tmp/context

echo '{"Ref": "'${SERVER}${DOCKER_USERNAME}'/'${FN}':latest"}' > tmp/com.openfaas.docker.config

cp -r build/$FN/* tmp/context
tar -C ./tmp -cvf $FN-context.tar .
```

## (Or) Create your build context  (for ECR)

Build the function:

```sh
export DOCKER_USERNAME=""
export REGION="eu-central-1"
export ACCOUNT_ID="012345678900"

export SERVER="$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com/"

faas-cli build --shrinkwrap -f $FN.yml

rm -fr tmp
mkdir -p tmp/context

echo '{"Ref": "'${SERVER}''${FN}':latest"}' > tmp/com.openfaas.docker.config

cp -r build/$FN/* tmp/context
tar -C ./tmp -cvf $FN-context.tar .
```

Now find the Pod and copy in your `.aws/credentials` file

```
kubectl get pod -n openfaas | grep of-builder
of-builder-8678c95d6f-fnblk

kubectl cp ~/.aws/credentials -n openfaas of-builder-8678c95d6f-fnblk:/home/app/.aws/credentialss
```

Alternatively, you could create a Kubernetes secret and update the of-builder's Deployment to mount it.

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
