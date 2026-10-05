<div align="center">
<img src="./assets/logo_round.png" width="140" align="center" alt="KomaruGram">

# KomaruGram Go

Telegram Desktop client written from scratch in Go. Feature-rich and security-enhanced foundation. Easy to build and freely distributable, using a permissive license.

[ [English] | [Русский](README_RU.md) ]

[ [Other stuff](./docs/README_FULL.md) | [Widgets and UI](./docs/UI_COMPONENTS.md) ]

</div>

## Why does this exist?

The official Telegram Desktop and its forks have many issues: heavy reliance on the C++ codebase and libraries ([TDLib](https://github.com/tdlib/td), Qt), a build process that is time-consuming and resource-intensive, and a high barrier to entry for developing your own modifications. In addition, it uses a copyleft license — our repository is completely free in terms of how you distribute it.

## What's inside?

[Gio](https://gioui.org/) handles the UI. It is a cross-platform, immediate-mode GUI library. It was chosen because it eliminates the typical limitations of frameworks, has no dependencies, doesn't break cross-platform builds, is fully compatible with Wayland, and can be ported to other platforms without any issues.

[gotd/td](https://github.com/gotd/td) implements MTProto and a little more, which serves as the foundation for all interactions with the Telegram protocol.

[wazero](https://github.com/wazero/wazero) is used as a high-performance WebAssembly sandbox for rendering stickers. It ensures cross-platform portability and guarantees code isolation without launching a separate process.

FFmpeg, mpv, VLC and Chromium are available as external integrations that you must provide yourself. However, the application will work even without them, if the user is satisfied with that.

Without FFmpeg, or when chosen in the settings, GIFs and animated avatars play with FFmpeg's H.264 decoder compiled to WebAssembly. It is not built into the application: it is downloaded the first time it is needed from [libavcodec-wasm](https://github.com/komarugif/libavcodec-wasm), which holds its sources and build script. To use your own build, point the `KOMARUGRAM_AVCDEC` environment variable at an `avcdec.wasm` file or its URL.

Voice messages and music play in the application itself, with no FFmpeg or external player, as they download: Opus with libopus, MP3, FLAC and WAV with dr_libs, all compiled to WebAssembly and built in. AAC (M4A) plays with the Fraunhofer FDK AAC decoder, which is downloaded the first time it is needed from [fdk-aac-wasm](https://github.com/komarugif/fdk-aac-wasm), where its sources are; `KOMARUGRAM_AACDEC` points at your own build.

## How to get started

Make a copy of this repository. Ensure that **[Go 1.27.1](https://go.dev/dl/)** is installed on your machine — it is the minimum required version. On Linux, Wayland and X11 development libraries may be required.

If you do not need to create an executable file, you can use this command to run it universally across different operating systems.
```
go run ./cmd/messenger
```

If you need to create an executable file with an icon, proceed as follows.

### **Linux**:
**Simply with go build:**
```
GOOS=linux go build -ldflags="-s -w" -buildvcs=false -o komarugram ./cmd/messenger
```

### **Windows:**

**Install Gio cmd tools:**
```
go install gioui.org/cmd/gogio@latest
```
**Building the application with an icon:**
```
gogio -ldflags="-s -w" -icon=./assets/logo_round.png -target=windows -o komarugram.exe ./cmd/messenger
```

The build process can consume up to 4 GB of RAM. Keep this in mind and close unnecessary applications during the initial build. All subsequent builds should run instantly.

## Portability

KomaruGram Go can be ported to a wide range of operating systems thanks to its architecture and minimal dependencies. Check out [PLATFORMS.md](./docs/PLATFORMS.md) if you are interested in platform-specific details.

| OS | Status |
|---|---|
| Windows 10/11 | ✅ First-class support |
| Linux (Wayland) | ✅ First-class support |
| Linux (X11) | ✅ Supported, within X11 limitations |
| MacOS | ✅ First-class support |
| FreeBSD | ⚠️ Should build; not tested in practice |
| OpenBSD | ⚠️ Should build in theory; not tested |
| NetBSD | ❓ Porting possible, with some caveats |
| Android | ❓ Porting possible, with some caveats |
| Windows 7 | ❓ Porting possible, with some caveats |
| Haiku OS | ❓ Porting possible, with significant caveats |

## Vibecoding

The code in this project was primarily written by the GPT-6 Astra and Claude Opus 5.5 language models. The maintainer is responsible for the concept, selection of the technology stack and libraries, and quality control.

## Thanks

We thank the creator of [gotd/td](https://github.com/gotd/td) for the excellent library and [Gio](https://gioui.org/) for an architecture that is both simple and scalable for large applications. Branding is provided free of charge by the [t.me/komarugram](https://t.me/komarugram) project. Thanks augustwise and SvatoshGPT for providing the Claude Code and Codex subscriptions for the needs of this project.

## You Might Be Interested In

If you were looking for the KomaruGram that is a fork of Telegram Desktop, [click here](https://github.com/svatoshgpt/komarugram). Our project has similar goals, but it was written from scratch in Go and is not a fork of Telegram Desktop

## License

When working with this project, you have no licensing obligations regarding the modification or distribution of the code. Some dependencies require that you respect copyright, but they do not impose any restrictions on the code itself (permissive MIT-compatible licenses).

There are two exceptions, both kept out of this repository and out of the binary, loaded at run time, and easy to replace or leave out:

- The H.264 decoder is FFmpeg's, licensed under the LGPL 2.1 or later. It is in [libavcodec-wasm](https://github.com/komarugif/libavcodec-wasm); `KOMARUGRAM_AVCDEC` replaces it.
- The AAC decoder is the Fraunhofer FDK AAC Codec Library, under Fraunhofer's own license: free to copy and distribute, with its source, but granting no patent rights. AAC is covered by patents licensed through [Via LA](https://www.via-la.com): using the decoder may need such a license, in particular for commercial use. Debian counts it as non-free for this reason. It is in [fdk-aac-wasm](https://github.com/komarugif/fdk-aac-wasm); `KOMARUGRAM_AACDEC` replaces it, and without it only M4A sound does not play.

## Third-party licenses

The Go modules linked into `messenger` and `kitchen`, as `go version -m` lists them for Linux and Windows builds:

| Component | License |
|---|---|
| Go (standard library and runtime) | BSD-3-Clause |
| [Gio](https://gioui.org) (`third_party/gio`, `gioui.org/shader`) | Unlicense OR MIT |
| [gio-mw](https://git.sr.ht/~schnwalter/gio-mw) (`third_party/gio-mw`) | Unlicense OR MIT |
| [gotd/td](https://github.com/gotd/td) | MIT |
| [gotd/ige](https://github.com/gotd/ige) | MIT |
| [gotd/log](https://github.com/gotd/log) | Apache-2.0 |
| [gotd/neo](https://github.com/gotd/neo) | BSD-3-Clause |
| [go-faster/errors](https://github.com/go-faster/errors), [go-faster/xor](https://github.com/go-faster/xor) | BSD-3-Clause |
| [go-faster/jx](https://github.com/go-faster/jx) | MIT |
| [wazero](https://github.com/tetratelabs/wazero) | Apache-2.0 |
| [ebitengine/oto](https://github.com/ebitengine/oto), [ebitengine/purego](https://github.com/ebitengine/purego) | Apache-2.0 |
| [jfreymuth/pulse](https://github.com/jfreymuth/pulse) | MIT |
| [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3), [ncruces/julianday](https://github.com/ncruces/julianday) | MIT |
| [ncruces/go-sqlite3-wasm](https://github.com/ncruces/go-sqlite3-wasm) | MIT-0; SQLite inside it is public domain |
| [lukechampine.com/adiantum](https://github.com/lukechampine/adiantum) | MIT |
| [go-text/typesetting](https://github.com/go-text/typesetting) | Unlicense OR BSD-3-Clause |
| [godbus/dbus](https://github.com/godbus/dbus) | BSD-2-Clause |
| [google/go-tpm](https://github.com/google/go-tpm) | Apache-2.0 |
| [coder/websocket](https://github.com/coder/websocket) | ISC |
| [refraction-networking/utls](https://github.com/refraction-networking/utls) | BSD-3-Clause |
| [klauspost/compress](https://github.com/klauspost/compress) | BSD-3-Clause |
| [andybalholm/brotli](https://github.com/andybalholm/brotli) | MIT |
| [cespare/xxhash](https://github.com/cespare/xxhash) | MIT |
| [segmentio/asm](https://github.com/segmentio/asm) | MIT-0 |
| [cenkalti/backoff](https://github.com/cenkalti/backoff) | MIT |
| [srwiley/oksvg](https://github.com/srwiley/oksvg), [srwiley/rasterx](https://github.com/srwiley/rasterx) | BSD-3-Clause |
| [uber-go/zap](https://github.com/uber-go/zap), [multierr](https://github.com/uber-go/multierr), [atomic](https://github.com/uber-go/atomic) | MIT |
| [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go) (`otel`, `otel/trace`) | Apache-2.0 |
| `golang.org/x/crypto`, `exp`, `exp/shiny`, `image`, `net`, `sync`, `sys`, `text` | BSD-3-Clause |
| [rsc.io/qr](https://github.com/rsc/qr) | BSD-3-Clause |
| [Go fonts](https://go.dev/blog/go-fonts) (`gioui.org/font/gofont`) | BSD-3-Clause |

Embedded and downloaded files:

| Component | License |
|---|---|
| `pkg/vp9/vpxdec.wasm`: [libvpx](https://github.com/webmproject/libvpx) with [SIMDe](https://github.com/simd-everywhere/simde) | BSD-3-Clause (libvpx), MIT (SIMDe) |
| `pkg/opus/opusdec.wasm`: [libopus](https://github.com/xiph/opus) (notice in `pkg/opus/COPYING.libopus`) | BSD-3-Clause |
| `pkg/drdec/drdec.wasm`: [dr_libs](https://github.com/mackron/dr_libs) (dr_mp3, dr_flac, dr_wav; `pkg/drdec/LICENSE.dr_libs`) | Unlicense OR MIT-0 |
| `pkg/lottie/tlottie.wasm`: [tlottie](https://github.com/dkaraush/tlottie) and the Rust standard library | MIT; MIT OR Apache-2.0 |
| WASI libc and compiler-rt inside the modules above ([wasi-sdk](https://github.com/WebAssembly/wasi-sdk)) | MIT, Apache-2.0 WITH LLVM-exception |
| `pkg/miniapp/assets/telegram-web-app.js`: Telegram's Mini App SDK | © Telegram, no license stated |
| `avcdec.wasm`, downloaded at run time: FFmpeg's H.264 decoder ([libavcodec-wasm](https://github.com/komarugif/libavcodec-wasm)) | LGPL-2.1-or-later |
| `aacdec.wasm`, downloaded at run time: the Fraunhofer FDK AAC decoder ([fdk-aac-wasm](https://github.com/komarugif/fdk-aac-wasm)) | Fraunhofer FDK AAC license, no patent grant |

The full texts are in each module's `LICENSE` file (`go env GOMODCACHE`) and in the repositories linked above.
