
## Install with Helm

You will need to clone the repo : 
```sh
git clone https://github.com/ddvk/rmfakecloud && cd rmfakecloud/helm

Modify your variables just like a simple docker-compose
```sh
vim values.yml
```

Install the helm chart
```sh
helm install myrmfakecloud .
```

## Create first user

Create your first user using kubectl exec:
```sh
export POD_NAME=$(kubectl get pods --namespace {{ .Release.Namespace }} -l "app.kubernetes.io/name={{ include "rmfakecloud.name" . }},app.kubernetes.io/instance={{ .Release.Name }}" -o jsonpath="{.items[0].metadata.name}")
kubectl exec $POD_NAME -- /rmfakecloud-docker setuser -u ddvk -a
```
(You may need to restart the pod)

You can reset a password with : 
```sh
export POD_NAME=$(kubectl get pods --namespace {{ .Release.Namespace }} -l "app.kubernetes.io/name={{ include "rmfakecloud.name" . }},app.kubernetes.io/instance={{ .Release.Name }}" -o jsonpath="{.items[0].metadata.name}")
kubectl exec $POD_NAME -- /rmfakecloud-docker setuser -u ddvk -p "${NEWPASSWD}"
```

