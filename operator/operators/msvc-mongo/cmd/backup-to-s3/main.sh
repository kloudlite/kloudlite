#! /usr/bin/env bash

set -o nounset
set -o pipefail
set -o errexit

encryption_key=$BACKUP_ENCRYPTION_KEY
s3_endpoint=$S3_ENDPOINT
s3_bucket_name=$S3_BUCKET_NAME
s3_path=$S3_PATH

# #!/usr/bin/env zx
#
# import {getEnv} from "./helpers/env.mjs"
#
# $.verbose = true
#
# async function main() {
#   const mongodbUri = getEnv("MONGODB_URI")
#
#   // Data leave empty for all databases
#   const dbs = getEnv("MONGODB_DATABASES", "")
#   const databasesToBackup = dbs ? dbs.split(","): []
#
#   const s3Endpoint = getEnv("S3_ENDPOINT")
#   const s3BucketName = getEnv("S3_BUCKET_NAME")
#   const s3Path = getEnv("S3_PATH")
#
#   getEnv("AWS_ACCESS_KEY_ID")
#   getEnv("AWS_SECRET_ACCESS_KEY")
#
#   // Backup configuration
#   const backupPath = (await $`mktemp -d`).toString().trim();
#
#   const encryptionKey = process.env.BACKUP_ENCRYPTION_KEY;
#
#   const date = new Date().toISOString().replace(/[:\-]|\.\d{3}/g, '');
#   const backupName = `mongodb_backup_${date}.tar.gz`;
#
#   const dbArgs = databasesToBackup.map(db => `--db ${db}`).join(" ")
#
#   // Create backup
#   console.log("Creating MongoDB backup...");
#   console.log("databasesToBackup", databasesToBackup)
#
#   const cmd = `mongodump --uri ${mongodbUri} ${dbArgs} --archive=${path.join(backupPath, backupName)} --dumpDbUsersAndRoles --gzip`
#
#   await $`${cmd.split(" ")}`;
#
#   // Encrypt the backup
#   console.log("Encrypting backup...");
#   await $`openssl enc -pbkdf2 -salt -in ${path.join(backupPath, backupName)} -out ${path.join(backupPath, `${backupName}.enc`)} -pass pass:${encryptionKey}`;
#
#   // Upload to S3
#   console.log("Uploading backup to S3...");
#   await $`s5cmd --endpoint-url=${s3Endpoint} cp ${path.join(backupPath, `${backupName}.enc`)} s3://${s3BucketName}/${s3Path}/${backupName}.enc`;
#
#
#   // Clean up
#   console.log("Cleaning up local backup files...");
#   await fs.promises.unlink(path.join(backupPath, backupName));
#   await fs.promises.unlink(path.join(backupPath, `${backupName}.enc`));
#
#   console.log("Backup process completed.");
# }
#
# try {
#   await main()
# } catch (err) {
#   console.error(err);
#   process.exit(1);
# }
