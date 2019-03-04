#!/bin/sh

DIR="$( dirname "$(readlink -f "$0")" )"

~/go/bin/mockgen -destination $DIR/mock_nakama.go -package mocks github.com/heroiclabs/nakama/runtime Logger
