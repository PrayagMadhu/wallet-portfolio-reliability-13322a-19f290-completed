#!/usr/bin/env bash
set -u

ROOT="/root/task"

echo "== Go dependency download =="
cd "$ROOT/api" || exit 2
go mod download
if [ $? -ne 0 ]; then
  echo "go mod download failed"
  exit 2
fi

echo "== Go compile/package smoke =="
go test -run '^$' ./...
if [ $? -ne 0 ]; then
  echo "go package loading failed"
  exit 2
fi

echo "== Go full test suite =="
go test ./...
GO_STATUS=$?
if [ "$GO_STATUS" -gt 1 ]; then
  echo "go test ended unexpectedly with status $GO_STATUS"
  exit 2
fi

echo "== Web dependency install =="
cd "$ROOT/web" || exit 2
npm install --silent
if [ $? -ne 0 ]; then
  echo "npm install failed"
  exit 2
fi

echo "== Web build smoke =="
npm run build --silent
if [ $? -ne 0 ]; then
  echo "web build failed"
  exit 2
fi

echo "== Web test suite =="
npm run test --silent -- --run
WEB_STATUS=$?
if [ "$WEB_STATUS" -gt 1 ]; then
  echo "web tests ended unexpectedly with status $WEB_STATUS"
  exit 2
fi

echo "Deployability probe completed. Behavioral test failures are expected in the starter."
exit 0
