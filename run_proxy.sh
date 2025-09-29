#!/bin/bash

set -e

direnv exec . go run cmd/codex-proxy/main.go
