```
# setup
docker rm -f of-builder
docker service rm registry

docker service create --network func_functions --name registry --detach=true  registry:latest
docker build -t alexellis2/of-builder .

docker run -d --net func_functions --name of-builder --privileged alexellis2/of-builder

rm req.tar

# prepare request tar
echo >config<<EOT                                                                                                        
{"Ref": "registry.local:5000/foo/bar:latest"}
EOT

mkdir -p context

cat >context/Dockerfile<<EOT                                                                                            
FROM busybox
ADD Dockerfile /
ENV foo bar
EOT

tar cvf req.tar .

# query
curl -i 192.168.10.98:8080/build -X POST --data-binary @req.tar
```