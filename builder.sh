#!/bin/bash
mkdir -p execs
for arch in amd64 arm64; do
  for os in windows darwin linux; do
    GOOS=$os GOARCH=$arch go build -ldflags="-s -w" -o=execs/"3xp1-$os-$arch"
    chmod +x execs/*
  done
done
