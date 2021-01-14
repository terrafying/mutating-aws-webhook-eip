./deployment/webhook-create-signed-cert.sh --namespace byoip
cat ./deployment/validatingwebhook.yaml | ./deployment/webhook-patch-ca-bundle.sh > ./deployment/validatingwebhook-ca-bundle.yaml
kubectl label namespace default byoip-mutator=enabled
cat ./deployment/mutatingwebhook.yaml | ./deployment/webhook-patch-ca-bundle.sh > ./deployment/mutatingwebhook-ca-bundle.yaml
# kubectl create -f deployment/validatingwebhook-ca-bundle.yaml
kubectl apply -f deployment/deployment.yaml
kubectl apply -f deployment/service.yaml
kubectl apply -f ./deployment/mutatingwebhook-ca-bundle.yaml
