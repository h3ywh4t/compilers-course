#!/usr/bin/env bash
set -e

go run ./cmd/compilerlab --file ./examples/err_semantic_type.cl || true
go run ./cmd/compilerlab --file ./examples/err_runtime_index.cl || true