#!/bin/bash
mkdir -p .github/workflows

# 1. DevSecOps CI/CD Pipeline
cat << 'YML' > .github/workflows/devsecops.yml
name: DevSecOps Pipeline

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

jobs:
  sast:
    name: Security Scan (gosec)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Gosec Security Scanner
        uses: securego/gosec@master
        with:
          args: ./...

  lint:
    name: Code Quality (golangci-lint)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v1.56.2

  test:
    name: Build, Test & Race Detection
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Build
        run: go build -v ./...
      - name: Test
        run: go test -v -race -coverprofile=coverage.out ./...
YML

# 2. Dependabot Configuration
cat << 'YML' > .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
YML

# 3. Security Policy
cat << 'MD' > SECURITY.md
# Security Policy

## Supported Versions

Only the latest `main` branch and tagged releases are currently receiving security updates.

## Reporting a Vulnerability

We take the security of our software very seriously. If you discover a vulnerability, we kindly request that you immediately report it. 

**Do not open a public issue.** Instead, please send an email to the repository owner or use the GitHub Security Advisory "Report a vulnerability" feature on this repository.

Please include:
* A detailed description of the vulnerability.
* Steps to reproduce the issue.
* Any potential mitigations you might propose.

We will endeavor to respond to your report within 48 hours and provide a timeline for a fix.
MD

# 4. Makefile for local DevSecOps
cat << 'MAKE' > Makefile
.PHONY: all build test clean sec lint bench

all: lint sec test build

build:
	go build -v ./...

test:
	go test -v -race ./...

bench:
	go run benchmark/main.go

sec:
	# Requires gosec: go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...

lint:
	# Requires golangci-lint
	golangci-lint run ./...

clean:
	rm -f cpu.prof mem.prof vp8_benchmark_results.csv coverage.out
	go clean
MAKE

chmod +x setup_devsecops.sh
./setup_devsecops.sh
rm setup_devsecops.sh
