#! /usr/bin/env bash

role="kloudlite_custom_role"
project_id="rich-wavelet-412321"

service_account="kloudlite-sa"
# export CLOUDSDK_AUTH_CREDENTIAL_FILE_OVERRIDE="/home/nxtcoder17/Downloads/rich-wavelet-412321-adc26c13a544.json"
# export CLOUDSDK_AUTH_CREDENTIAL_FILE_OVERRIDE="/tmp/gcp-auth.json"

# gcloud auth login

gcloud iam roles create $role \
	--project=$project_id \
	--title="Kloudlite Role" \
	--description="kloudlite admin role" \
	--permissions=compute.instances.start,compute.instances.stop \
	--stage=GA | tee -a /tmp/output.json

gcloud iam service-accounts create $service_account \
	--description="service account json" \
	--display-name="Kloudlite Account" \
	--project="$project_id"

gcloud projects add-iam-policy-binding "$project_id" \
	--member="serviceAccount:$service_account@" \
	--role="projects/$project_id/roles/$role"
