# JUMON.md

A magically simple AI workflow orchestration engine.

Status: In development, specifications are subject to change.

## Overview

JUMON is a powerful yet simple AI workflow orchestration engine that allows you to define, manage, and execute AI workflows using Markdown files. It provides a declarative way to create AI-powered applications by combining scripts, tools, and AI models in a modular fashion.

## Features

- **Markdown-Based Workflows**: Define your AI workflows in simple JUMON.md files
- **Modular Architecture**: Organize your code into reusable modules
- **Multi-Model Support**: Integrate with various LLM providers (OpenAI, Anthropic, Google, Ollama)
- **Tool Integration**: Extend functionality with custom tools (WASM, NATS, scripts)
- **Script Orchestration**: Create sequences of AI-powered steps
- **Event System**: React to events with custom handlers

## Installation

```bash
# Install JUMON
go install github.com/jumonmd/jumon@latest

# Verify installation
jumon version
```

## Quick Start

1. Initialize a new JUMON module:

```bash
jumon init hello-world
```

2. Edit the generated JUMON.md file:

```markdown
---
module: hello-world
---

## Scripts

### main

1. How is JUMON written in kanji?
2. Explain the meaning along with its kanji
```

3. Run your module:

```bash
jumon run .
```

## Usage

```
jumon run <url_or_path> <input>
```

or start jumon server separately

```
jumon serve
```

### Commands

- `jumon serve`: Start the JUMON server
- `jumon stop`: Stop the JUMON server
- `jumon init <name>`: Initialize a new JUMON module
- `jumon run <url_or_path> [input]`: Run a JUMON module
- `jumon version`: Show the version

## Documentation

For more detailed documentation, see the [JUMON.md/docs](https://JUMON.md/docs).

## Telemetry

Anonymous telemetry is enabled by default for quality improvement purposes, collecting usage counts and error codes along with version and OS environment information. You can disable it using the following method:

```
jumon --disable-telemetry
```


## Tasks

### test

```
go test -v -failfast ./chat/... || exit 1
go test -v -failfast ./event/... || exit 1
go test -v -failfast ./internal/... || exit 1
go test -v -failfast ./module/... || exit 1
go test -v -failfast ./script/... || exit 1
go test -v -failfast ./tool/... || exit 1
```

### lint

```
golangci-lint run
```

### dist

Requires: test, lint

```
goreleaser release --snapshot --clean
```

## License

[MPL-2.0](LICENSE)
