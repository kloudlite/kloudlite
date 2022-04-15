SECRET=$1
IP=$2
#k3s server --server https://${IP}:6443 --node-name=secondary --token=${SECRET} 
k3s server --server https://${IP}:6443 --cluster-init --disable=traefik --token=${SECRET}
