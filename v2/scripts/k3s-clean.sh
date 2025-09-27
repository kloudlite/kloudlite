#!/bin/bash

# Clean all Kloudlite CRDs and resources

echo "Cleaning Kloudlite resources..."

# Delete all custom resources first
for crd in $(docker exec kloudlite-k3s kubectl get crd -o name | grep -E "platform.kloudlite.io|auth.kloudlite.io"); do
  resource=$(echo $crd | cut -d'/' -f2 | cut -d'.' -f1)
  echo "Deleting all $resource resources..."
  docker exec kloudlite-k3s kubectl delete $resource --all -A 2>/dev/null || true
done

# Delete CRDs
echo "Deleting CRDs..."
docker exec kloudlite-k3s kubectl delete crd -l app.kubernetes.io/name=kloudlite 2>/dev/null || true

echo "Cleanup complete"