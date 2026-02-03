#!/bin/bash

GOOS=js GOARCH=wasm go build -o main.wasm entry/js/main.go