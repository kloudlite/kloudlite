# KUBECONFIG=$1

# CONFIG_DIR=/Users/abdeshnayak/kloudlite/api-go/apps/infra/internal/application/tf/do/init-scripts/ingress/
# CONFIG_DIR=/Users/abdeshnayak/kloudlite/api-go/apps/infra/internal/application/tf/do/init-scripts/ingress/
CONFIG_DIR=.

kubectl apply -f $CONFIG_DIR/csi/crds.yaml

helm install $CONFIG_DIR/ingress

kubectl apply -f $CONFIG_DIR/csi 

kubectl apply -f $CONFIG_DIR/wireguard

