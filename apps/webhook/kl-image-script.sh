#!/bin/bash

meta=$(echo "$meta" | tr -d ' ')

json_data='{"image": "'"$image"'", "meta": {'

IFS=',' read -r -a array <<< "$meta"
for element in "${array[@]}"
do
  key=$(echo "$element" | cut -d '=' -f 1)
  value=$(echo "$element" | cut -d '=' -f 2)
  json_data+='"'"$key"'":"'"$value"'",'
done

json_data=${json_data%,}'}}'

curl -X POST $WEBHOOK_URL -H "Authorization: $authorization" -H "Content-Type: application/json" -d "$json_data"
