#!/usr/bin/env bash

declare -i last_called=0
declare -i throttle_by=2

@throttle() {
  local -i now=$(date +%s)
  if (($now - $last_called >= $throttle_by))
  then
    "$@"
  fi
  last_called=$(date +%s)
}


@debounce() {
  if [[ ! -f ./executing ]]
  then
    touch ./executing
    "$@"
    retVal=$?
    {
      sleep $throttle_by
      if [[ -f ./on-finish ]]
      then
        "$@"
        rm -f ./on-finish
      fi
      rm -f ./executing
    } &
    return $retVal
  elif [[ ! -f ./on-finish ]]
  then
    touch ./on-finish
  fi
}

@debounce make dev
wait $(jobs -p) # need to wait for the bg jobs to complete
