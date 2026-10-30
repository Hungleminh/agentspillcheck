# AgentSpillCheck

Tiếng Việt · [English](README.md)

![license](https://img.shields.io/badge/license-Apache--2.0-blue) ![telemetry](https://img.shields.io/badge/telemetry-none-green)

### Video demo

[![Xem demo 140s — thuyết minh tiếng Anh có phụ đề](video/out/demo-poster.jpg)](https://youtu.be/V8oSr_aFQNM)

**Đưa dữ liệu giả cho AI agent, rồi biết nó đã đọc gì, mang gì đi đâu, giữ lại ở đâu, và có mã hóa hay không — mà không chạm vào một byte dữ liệu thật nào của bạn.**

## Vì sao cần

- Bạn đã cho AI agent đọc email, mở file, thậm chí giữ mật khẩu.
- Bạn không biết nó mang những thứ đó đi đâu, có in ra chỗ không nên, có gửi tới bên thứ ba, có bị lệnh ẩn dẫn dắt hay không.
- Công cụ kiểm tra hiện có dành cho chuyên gia bảo mật. AgentSpillCheck dành cho người dùng bình thường: bấm một lệnh, đọc kết quả bằng tiếng Việt.

## Chạy thử 60 giây

```sh
spillcheck demo
```

Lệnh này chạy một agent giả cố tình làm lộ dữ liệu và in hạng/điểm cùng số phát hiện.

## Cài đặt

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

## Nối agent của bạn

In đoạn cấu hình MCP để dán vào Claude Desktop, Cursor hoặc Claude Code:

```sh
spillcheck mcp config --client claude-desktop
spillcheck mcp serve --stdio
```

Agent tự nhận việc qua `spill_start`/`spill_next_task`, làm việc bằng các tool `stage_*`, rồi `spill_finish`. Cũng có REST API tại `http://127.0.0.1:7700/api/v1` và connector cho Claude Code, agent HTTP tự viết, hoặc chế độ thủ công.

## Đọc báo cáo

Báo cáo trả lời bốn câu: đã đọc gì, có gửi đi không, gửi đến đâu, có mã hóa không. Mức độ: S0 an toàn trong phiên · S1 lộ với chính bạn · S2 giữ lại hoặc gửi nhầm người quen · S3 gửi ra bên ngoài · S4 bị điều khiển hoặc che giấu. Báo cáo có bản HTML, JSON và Markdown, không chứa giá trị mồi đầy đủ.

## Những gì công cụ không thấy được

Kết quả là bằng chứng của một phiên cụ thể, không phải chứng nhận an toàn. Nếu agent gửi dữ liệu qua đường không đi qua sân khấu giả hay proxy, công cụ không thấy — báo cáo luôn ghi rõ giới hạn này.

## Bảo mật và sử dụng có trách nhiệm

Chỉ kiểm thử agent bạn sở hữu hoặc được phép kiểm thử. Dữ liệu sinh ra là giả. Công cụ tự host, không telemetry, không gửi gì đi đâu trừ khi bạn bật rõ. Xem [SECURITY.md](SECURITY.md).

## Giấy phép

Giấy phép Apache-2.0 — xem [LICENSE](LICENSE) và [NOTICE](NOTICE).
