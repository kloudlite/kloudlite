
build:
	GOOS=linux GOARCH=amd64 go build -o bin/wgman

push:
	scp bin/wgman  root@143.110.242.223:/root/bin/wgman

test.vminit:
	ssh root@143.110.242.223 wgman -command=init -ip=143.110.242.223
	ssh root@143.110.242.223 wg

test.vmpeers: data=$(shell cat peers.json | base64)
test.vmpeers:
		ssh root@143.110.242.223 wgman -command=peers -peers=$(data)
		ssh root@143.110.242.223 wg

test.podinit:
	ssh root@143.110.242.223 k3s kubectl exec -n wireguard wireguard -- /host-scripts/wgman -command=init -ip=143.110.242.223
	ssh root@143.110.242.223 k3s kubectl exec -n wireguard wireguard -- wg

test.podpeers: data=$(shell cat peers.json | base64)
test.podpeers:
	ssh root@143.110.242.223 k3s kubectl exec -n wireguard wireguard -- /host-scripts/wgman -command=peers -peers=$(data)
	ssh root@143.110.242.223 k3s kubectl exec -n wireguard wireguard -- wg
