# KUBECONFIG=$1

# CONFIG_DIR=/Users/abdeshnayak/kloudlite/api-go/apps/infra/internal/application/tf/do/init-scripts/ingress/
# CONFIG_DIR=/Users/abdeshnayak/kloudlite/api-go/apps/infra/internal/application/tf/do/init-scripts/ingress/
CONFIG_DIR=.

sh $CONFIG_DIR/ingress/init.sh install
sh $CONFIG_DIR/cert-manager/install.sh install

kubectl apply -f $CONFIG_DIR/csi/crds.yaml
kubectl apply -f $CONFIG_DIR/csi
kubectl apply -f $CONFIG_DIR/wireguard
kubectl apply -f ./cert-manager/crd.yml
