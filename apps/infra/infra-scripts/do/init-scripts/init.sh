CONFIG_DIR=.

kubectl apply -f $CONFIG_DIR/wireguard
kubectl apply -f $CONFIG_DIR/csi/crds.yaml
kubectl apply -f $CONFIG_DIR/csi
kubectl apply -f $CONFIG_DIR/cert-manager/crd.yml
kubectl apply -f $CONFIG_DIR/knative
kubectl delete sc local-path



sh $CONFIG_DIR/loki/init.sh upgrade
sh $CONFIG_DIR/ingress/init.sh upgrade
sh $CONFIG_DIR/cert-manager/install.sh upgrade
sh $CONFIG_DIR/linkerd/install.sh upgrade


helm uninstall traefik --namespace kube-system
helm uninstall traefik-crd --namespace kube-system

