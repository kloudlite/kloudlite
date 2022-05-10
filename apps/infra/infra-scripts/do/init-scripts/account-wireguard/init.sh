cat <<EOF | kubectl create -f -
apiVersion: crds.kloudlite.io/v1
kind: Account
metadata:
  name: $1
EOF
