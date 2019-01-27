#!/bin/bash

DIR="$( dirname "$(readlink -f "$0")" )"

echo "Building nakama go module"
docker run --rm -v "$DIR/build:/build" -v "$DIR:/go/src/github.com/mastern2k3/poseidon" heroiclabs/nakama-pluginbuilder:2.3.2 build --buildmode=plugin -o /build/poseidon_plugin.so /go/src/github.com/mastern2k3/poseidon/poseidon_plugin.go

if [[ $? -ne 0 ]]; then
    echo "Build failed"
    exit 1
fi

echo "Build finished -> $DIR/build"
