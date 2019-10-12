#!/bin/bash
WD="$(
    cd "$(dirname "$0")"
    pwd -P
)"
docker build -t malt_app $WD
