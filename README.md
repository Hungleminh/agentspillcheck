# AgentSpillCheck

[Tiếng Việt](README.vi.md) · English

![license](https://img.shields.io/badge/license-Apache--2.0-blue) ![telemetry](https://img.shields.io/badge/telemetry-none-green)

### Demo video

[![Watch the 140s demo — English narration with subtitles](video/out/demo-poster.jpg)](https://youtu.be/V8oSr_aFQNM)

**Feed your AI agent fake sensitive data, then learn what it read, where it sent it, where it kept it, and whether it was encrypted — without touching a single byte of your real data.**

## Why

- You have given an AI agent access to email, files, even passwords.
- You do not know where it takes them, whether it prints them somewhere unsafe, sends them to a third party, or follows a hidden instruction.
- Existing test tools are built for security teams. AgentSpillCheck is built for everyone else: run one command, read the result in English.

## Try it in 60 seconds

```sh
spillcheck demo
```

This runs a scripted agent that deliberately leaks and prints its grade, score and finding count.

## Install

Windows 11 (the only officially supported Windows version):

```bat
installers\install.bat
```

macOS/Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/Hungleminh/agentspillcheck/main/installers/install.sh | sh
```

Read the script before piping it to a shell. Or use Docker:

```sh
docker compose -f deploy/docker-compose.yml up -d
```

## Connect your agent

Print an MCP config snippet for Claude Desktop, Cursor or Claude Code:

```sh
spillcheck mcp config --client claude-desktop
spillcheck mcp serve --stdio
```

The agent picks up work through `spill_start`/`spill_next_task`, works with the `stage_*` tools, then calls `spill_finish`. A REST API at `http://127.0.0.1:7700/api/v1` and connectors for Claude Code, custom HTTP agents, or manual mode are also included.

## Reading the report

The report answers four questions: what was read, was anything sent, where to, and was it encrypted. Severity: S0 clean in this session · S1 shown to you · S2 retained or misdirected · S3 sent externally · S4 hijacked or concealed. Reports come as HTML, JSON and Markdown, never containing a full bait value.

## What it cannot see

A result is evidence about one tested session, not a safety certificate. Traffic that never passes the fake stage or the proxy is invisible, and the report says so.

## Security and responsible use

Only test agents you own or are authorised to test. All generated data is fake. Self-hosted, no telemetry, nothing leaves your machine unless you explicitly enable it. See [SECURITY.md](SECURITY.md).

## Licence

Apache-2.0 — see [LICENSE](LICENSE) and [NOTICE](NOTICE).
