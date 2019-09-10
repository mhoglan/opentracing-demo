#!/bin/bash

API_DEFINITION=$1
GATEWAY_URL=$2

if [ -n $GATEWAY_URL ]; then
    GATEWAY_URL="http://localhost:8181"
fi


curl -H "x-tyk-authorization: 352d20ee67be67f6340b4c0605b044b7" \
 -H "Content-Type: application/json" \
 -X POST \
 -d "@$1" $GATEWAY_URL/tyk/apis/


curl -H "x-tyk-authorization: 352d20ee67be67f6340b4c0605b044b7" -s $GATEWAY_URL/tyk/reload/group