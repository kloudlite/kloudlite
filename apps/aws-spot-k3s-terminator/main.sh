#! /usr/bin/env bash

node_name="${NODE_NAME}"
[ -z "${node_name}" ] && echo "env var 'NODE_NAME' is not set, exiting." && exit 1

debug_mode=${DEBUG}

url="http://169.254.169.254/latest/meta-data/spot/instance-action"

echo "running ..."

function debug_msg() {
	if [ "${debug_mode}" == "true" ]; then
		echo "[debug] " "$@"
	fi
}

while true; do
	debug_msg "executing 'curl --connect-timeout -s -f $url'"
	d=$(curl --connect-timeout 1 -s -f "$url")
	exit_code=$?
	debug_msg "exit_code is $exit_code (!= 0), so trying again in 3 seconds"
	if [ "$exit_code" -eq 0 ]; then
		echo "Instance is marked for termination: $d"
		term_time=$(echo -n "$d" | jq -r '.time')
		term_timestamp=$(date -d "$term_time" +%s)
		curr_timestamp=$(date +%s)

		diff=$((term_timestamp - curr_timestamp))
		echo "we have ${diff}s to drain and terminate the node ${node_name}"
		if [ $((diff)) -ge 0 ]; then
			debug_msg kubectl drain --ignore-daemonsets --delete-local-data --force --grace-period=$((diff - 10)) "${node_name}"
			kubectl drain --ignore-daemonsets --delete-local-data --force --grace-period=$((diff - 10)) "${node_name}"
			sleep $((diff))s
		fi
	else
		# need to uncordon if new spot node has arrived, and is ready
		node_status=$(kubectl get node/"${node_name}" -o jsonpath='{.status.conditions[?(.type == "Ready")].status}')
		node_is_schedulable=$(kubectl get node/"${node_name}" -o jsonpath='{.spec.unschedulable}')

		if [ "$node_status" = "True" ]; then
			[ "$node_is_schedulable" = "true" ] && echo "node ${node_name} is ready, but marked as unschedulable, uncordoning ..." && kubectl uncordon "${node_name}" && exit 0
		fi
	fi
	sleep 3
done
