#!/bin/bash

TOKEN=$(jq -r .access_token $1)
curl -s -L -H "Authorization: bearer $TOKEN" "https://graph.microsoft.com/v1.0$2"
