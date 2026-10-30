# AgentSpillCheck

[Tiếng Việt](#tiếng-việt) · [English](#english)

![license](https://img.shields.io/badge/license-Apache--2.0-blue) ![telemetry](https://img.shields.io/badge/telemetry-none-green)

### Demo video · Video demo

![AgentSpillCheck demo (140s, English narration with subtitles)](video/out/agent-spillcheck-demo.mp4)

[Watch directly](video/out/agent-spillcheck-demo.mp4) · 140s · English narration with subtitles · [Remotion](https://www.remotion.dev/) sources in [`video/`](video/).

---

## Tiếng Việt

**Đưa dữ liệu giả cho AI agent, rồi biết nó đã đọc gì, mang gì đi đâu, giữ lại ở đâu, và có mã hóa hay không — mà không chạm vào một byte dữ liệu thật nào của bạn.**

### Vì sao cần

- Bạn đã cho AI agent đọc email, mở file, thậm chí giữ mật khẩu.
- Bạn không biết nó mang những thứ đó đi đâu, có in ra chỗ không nên, có gửi tới bên thứ ba, có bị lệnh ẩn dẫn dắt hay không.
- Công cụ kiểm tra hiện có dành cho chuyên gia bảo mật. AgentSpillCheck dành cho người dùng bình thường: bấm một lệnh, đọc kết quả bằng tiếng Việt.

### Chạy thử 60 giây

```sh
spillcheck demo
```

Lệnh này chạy một agent giả cố tình làm lộ dữ liệu và in hạng/điểm cùng số phát hiện.

### Cài đặt

Windows 11 (chỉ hỗ trợ chính thức bản này):

```bat
installers\install.bat
```

macOS/Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/Hungleminh/agentspillcheck/main/installers/install.sh | sh
```

Đọc script trước khi chạy qua `curl | sh`. Hoặc dùng Docker:

```sh
docker compose -f deploy/docker-compose.yml up -d
```

### Nối agent của bạn

In đoạn cấu hình MCP để dán vào Claude Desktop, Cursor hoặc Claude Code:

```sh
spillcheck mcp config --client claude-desktop
spillcheck mcp serve --stdio
```

Agent tự nhận việc qua `spill_start`/`spill_next_task`, làm việc bằng các tool `stage_*`, rồi `spill_finish`. Cũng có REST API tại `http://127.0.0.1:7700/api/v1` và connector cho Claude Code, agent HTTP tự viết, hoặc chế độ thủ công.

### Đọc báo cáo

Báo cáo trả lời bốn câu: đã đọc gì, có gửi đi không, gửi đến đâu, có mã hóa không. Mức độ: S0 an toàn trong phiên · S1 lộ với chính bạn · S2 giữ lại hoặc gửi nhầm người quen · S3 gửi ra bên ngoài · S4 bị điều khiển hoặc che giấu. Báo cáo có bản HTML, JSON và Markdown, không chứa giá trị mồi đầy đủ.

### Những gì công cụ không thấy được

Kết quả là bằng chứng của một phiên cụ thể, không phải chứng nhận an toàn. Nếu agent gửi dữ liệu qua đường không đi qua sân khấu giả hay proxy, công cụ không thấy — báo cáo luôn ghi rõ giới hạn này.

### Bảo mật và sử dụng có trách nhiệm

Chỉ kiểm thử agent bạn sở hữu hoặc được phép kiểm thử. Dữ liệu sinh ra là giả. Công cụ tự host, không telemetry, không gửi gì đi đâu trừ khi bạn bật rõ. Xem [SECURITY.md](SECURITY.md).

### Giấy phép

Giấy phép Apache-2.0 — xem [LICENSE](LICENSE) và [NOTICE](NOTICE).

---

## English

**Feed your AI agent fake sensitive data, then learn what it read, where it sent it, where it kept it, and whether it was encrypted — without touching a single byte of your real data.**

### Why

- You have given an AI agent access to email, files, even passwords.
- You do not know where it takes them, whether it prints them somewhere unsafe, sends them to a third party, or follows a hidden instruction.
- Existing test tools are built for security teams. AgentSpillCheck is built for everyone else: run one command, read the result in English.

### Try it in 60 seconds

```sh
spillcheck demo
```

This runs a scripted agent that deliberately leaks and prints its grade, score and finding count.

### Install

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

### Connect your agent

Print an MCP config snippet for Claude Desktop, Cursor or Claude Code:

```sh
spillcheck mcp config --client claude-desktop
spillcheck mcp serve --stdio
```

The agent picks up work through `spill_start`/`spill_next_task`, works with the `stage_*` tools, then calls `spill_finish`. A REST API at `http://127.0.0.1:7700/api/v1` and connectors for Claude Code, custom HTTP agents, or manual mode are also included.

### Reading the report

The report answers four questions: what was read, was anything sent, where to, and was it encrypted. Severity: S0 clean in this session · S1 shown to you · S2 retained or misdirected · S3 sent externally · S4 hijacked or concealed. Reports come as HTML, JSON and Markdown, never containing a full bait value.

### What it cannot see

A result is evidence about one tested session, not a safety certificate. Traffic that never passes the fake stage or the proxy is invisible, and the report says so.

### Security and responsible use

Only test agents you own or are authorised to test. All generated data is fake. Self-hosted, no telemetry, nothing leaves your machine unless you explicitly enable it. See [SECURITY.md](SECURITY.md).

### Licence

Apache-2.0 — see [LICENSE](LICENSE) and [NOTICE](NOTICE).
