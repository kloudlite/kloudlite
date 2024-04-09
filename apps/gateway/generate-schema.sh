#! /usr/bin/env bash

out=$1

mkdir -p schemas

cat ../accounts/internal/app/graph/*.graphqls >./schemas/accounts-api.schema
cat ../accounts/internal/app/graph/struct-to-graphql/*.graphqls >>./schemas/accounts-api.schema

cat ../console/internal/app/graph/*.graphqls >./schemas/console-api.schema
cat ../console/internal/app/graph/struct-to-graphql/*.graphqls >>./schemas/console-api.schema

cat ../container-registry/internal/app/graph/*.graphqls >./schemas/container-registry-api.schema
cat ../container-registry/internal/app/graph/struct-to-graphql/*.graphqls >>./schemas/container-registry-api.schema

cat ../infra/internal/app/graph/*.graphqls >./schemas/infra-api.schema
cat ../infra/internal/app/graph/struct-to-graphql/*.graphqls >>./schemas/infra-api.schema

cat ../auth/internal/app/graph/*.graphqls >./schemas/auth-api.schema
cat ../message-office/internal/app/graph/*.graphqls >./schemas/message-office-api.schema

cat ../iot-console/internal/app/graph/*.graphqls >./schemas/iot-console-api.schema
cat ../iot-console/internal/app/graph/struct-to-graphql/*.graphqls >>./schemas/iot-console-api.schema

rover supergraph compose --config ./supergraph.yml --output "$out" --elv2-license accept
