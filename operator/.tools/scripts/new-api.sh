#! /usr/bin/env bash
operator-sdk create api --group crds --version v1 --resource --controller --kind $@ 
