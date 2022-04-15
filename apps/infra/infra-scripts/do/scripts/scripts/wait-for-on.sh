while sleep 1; 
do 
	res=$(k3s kubectl get nodes| grep -i AGE | wc -l) 
	if [[ 1 -eq $res ]]; then
		# k3s kubectl get nodes;
		break;
	fi
done

k3s kubectl wait --for=condition=Ready=True node/master-0 --timeout 7000s
