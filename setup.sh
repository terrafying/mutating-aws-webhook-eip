./deployment/webhook-create-signed-cert.sh
cat ./deployment/validatingwebhook.yaml | ./deployment/webhook-patch-ca-bundle.sh > ./deployment/validatingwebhook-ca-bundle.yaml
kubectl label namespace default admission-webhook-example=enabled
cat ./deployment/mutatingwebhook.yaml | ./deployment/webhook-patch-ca-bundle.sh > ./deployment/mutatingwebhook-ca-bundle.yaml
# kubectl create -f deployment/validatingwebhook-ca-bundle.yaml
kubectl create -f deployment/deployment.yaml
kubectl create -f ./deployment/mutatingwebhook-ca-bundle.yaml

