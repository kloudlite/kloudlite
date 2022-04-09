#/root/scripts/secondary/clean.sh
#/root/scripts/secondary/start.sh
#/root/scripts/wait-for-on.sh

SECRET=$1
IP=$2
k3s agent --server https://${IP}:6443 --token ${SECRET} --node-name agent-02


