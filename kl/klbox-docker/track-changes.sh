#! /usr/bin/env bash

filepath=$1
cmd=$2

LTIME=`stat -c %Z $filepath`
oldhash=$(jq -r '.hash' < $filepath)

while true
do
   ATIME=`stat -c %Z $filepath`

   if [[ "$ATIME" != "$LTIME" ]]; then
       newhash=$(jq -r '.hash' < "$filepath")

       if [ "$oldhash" != "$newhash" ]; then
          echo "[$(date)] file has changed" >> /tmp/track-changes.log
          eval $cmd
          LTIME=$ATIME
          oldhash=$newhash
       else
         echo "[$(date)] file hash has not changed" >> /tmp/track-changes.log
       fi

#   else
#     echo "[$(date)] file has not changed" >> /tmp/track-changes.log
   fi
   sleep 1
done