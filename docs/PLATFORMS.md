# Platforms: the graphics stack, and Unix systems other than Linux

How Gio draws on each platform, and what building KomaruGram for other
Unix-like systems takes, from a FreeBSD build tried in October 2026; and
what a port to Haiku would take, from its sources.

**What is checked stays Linux, Windows and macOS.** A native build for FreeBSD,
and likely OpenBSD, is possible and cheap to port, but nobody runs it:
live checks, the tests and the cross-builds in `AGENTS.md` are for Linux,
Windows and macOS only. A
change need not be checked on these systems, and nothing here is a
promise that they work.

## How Gio draws

Gio's renderer (`third_party/gio/gpu/gpu.go`) is one for every platform;
under it are the graphics APIs of each:

| Platform | Windows | Graphics | Chosen by |
|---|---|---|---|
| Linux, Wayland | `app/os_wayland.go` | OpenGL ES 3 through EGL; Vulkan if that fails | `NewContext`: EGL first |
| Linux, X11 | `app/os_x11.go`, `os_x11_xi2.go` | OpenGL ES 3 through EGL | Vulkan is there but off: `vulkanBuggy = true` |
| Windows | `app/os_windows.go` | Direct3D 11; EGL (ANGLE's `libEGL.dll`) if that fails | priority: D3D11 1, EGL 2 |
| macOS | `app/os_macos.go` | Metal; OpenGL with the `nometal` tag | — |
| Android, iOS, js | their own | GLES or Vulkan, Metal or GLES, WebGL | — |

- On Linux the client draws with OpenGL ES in practice. `libEGL` is
  linked; `libGLESv2.so.2` and `libvulkan.so.1` are loaded with `dlopen`
  at run time (`internal/gl/gl_unix.go`, `internal/vk/vulkan.go`).
- The blur of the overlays (`paint.PushBlur`) uses the existing blit
  shaders and adds none, so it works on every API.
- Render tests draw with `gpu/headless`, through EGL on Linux.
- The tags `nowayland`, `nox11`, `novulkan` and `noopengl` leave a backend
  out of the build.

## Gio on Unix systems other than Linux

- **FreeBSD**: X11 and Wayland, EGL and Vulkan. The cgo flags name
  FreeBSD's paths (`/usr/local/include`, `/usr/local/lib`) and link the
  libraries directly, without pkg-config.
- **OpenBSD**: X11 and EGL only, no Wayland or Vulkan.
- **NetBSD**: not supported by Gio's `app`; a build fails in
  `gioui.org/internal/gl`.
- Gio needs cgo on all of them: without it, `gioui.org/internal/vk` has no
  files to build.

## FreeBSD: built, not run

`cmd/messenger` was cross-built and linked for FreeBSD 14 on amd64 from
Linux: a dynamic ELF for FreeBSD 14.0 of 68 MB, needing only libraries of
the ports:

```
libwayland-egl.so.1  libwayland-client.so.0  libwayland-cursor.so.0
libX11.so.6  libxkbcommon.so.0  libxkbcommon-x11.so.0  libX11-xcb.so.1
libXcursor.so.1  libXfixes.so.3  libEGL.so.1  libthr.so.3  libc.so.7
```

`libGLESv2` (`mesa-libs`) and `libvulkan` are loaded at run time. It was
not run: the development machine is a VirtualBox guest without nested
virtualization, where FreeBSD with a desktop would run under emulation
alone. A real machine or a VM of FreeBSD is needed to try it.

### What stands in the way

1. **`pkg/deviceinfo/libc_linux_cgo.go`** is `//go:build cgo` and is built
   on every system with cgo: Go takes an OS from the last element of a
   file name only, and this one ends in `_cgo`. It includes glibc's
   `<features.h>`, which FreeBSD lacks; by the code macOS should fail the
   same way (not tried). The fix: `//go:build linux && cgo` and
   `linux && !cgo` for the two files, and a `libcVersion` for other
   systems. Not fixed yet; the build above replaced the files with
   `go build -overlay`.
2. **Vulkan's headers** are needed to build (`vulkan-headers` in the
   ports), or the `novulkan` tag.
3. **Memory**: compiling `github.com/gotd/td/tg` for a new target peaks at
   4.8 GB. On a machine with 6 GB, build it alone first
   (`go build -p 1 github.com/gotd/td/tg`), then the client.

### What works differently there

Read from the code, not tried:

| What | On FreeBSD |
|---|---|
| Tray | SNI over D-Bus, as on Linux (`internal/tray/tray_sni.go` names FreeBSD) |
| Sound (`oto`) | PulseAudio's protocol in pure Go, else ALSA loaded with `purego`: needs PulseAudio or PipeWire running, or `alsa-lib` with its OSS plugin |
| SQLite, the wasm sandbox | pure Go (`wazero`), as everywhere |
| TPM | `go-tpm` opens `/dev/tpmrm0`, and `/dev/tpm0` when that does not exist. FreeBSD's TPM 2.0 driver makes `/dev/tpm0` only (`TPM_CDEV_NAME` in [`sys/dev/tpm/tpm20.h`](https://github.com/freebsd/freebsd-src/blob/releng/14.3/sys/dev/tpm/tpm20.h)), without a resource manager. The diagnosis of access (`security/access_linux.go`) is Linux's only |
| Voice messages | not available: `pkg/voice/input_other.go` returns an error. FreeBSD's ffmpeg has `-f oss` and usually `-f pulse`; an `input_freebsd.go` would be a few lines |
| Dark theme, power saving | stubs (`appearance_other.go`, `powersave_other.go`), though the freedesktop portal and D-Bus are there on FreeBSD too |
| Giving memory back | no `malloc_trim` (glibc's) |

### How it was built

From Linux, with [zig](https://ziglang.org/download/) 0.16 as the C
compiler (it carries FreeBSD's libc headers and stubs) and the headers and
libraries of FreeBSD's packages, from
[pkg.freebsd.org](https://pkg.freebsd.org/FreeBSD:14:amd64/latest/):
`libX11`, `xorgproto`, `libxkbcommon`, `libXcursor`, `libXfixes`,
`libXrender`, `libxcb`, `libXau`, `libXdmcp`, `libXext`, `wayland`,
`libglvnd`, `libffi`, `libxml2`, `libepoll-shim` and `vulkan-headers`.
Their `.pkg` files are tar archives in zstd; the package catalog,
`packagesite.pkg`, gives the path of each (`repopath`). Unpacked into
`$ROOT`:

```sh
cat > cc <<EOF
#!/bin/sh
exec /path/to/zig cc -target x86_64-freebsd "\$@"
EOF
chmod +x cc
CGO_ENABLED=1 GOOS=freebsd GOARCH=amd64 CC=$PWD/cc \
  CGO_CFLAGS="-I$ROOT/usr/local/include" \
  CGO_LDFLAGS="-L$ROOT/usr/local/lib" \
  go build -p 2 -o messenger-freebsd ./cmd/messenger
```

zig, the packages and the build take about 1.5 GB of disk with the Go
build cache for the new target. On FreeBSD itself it is the ordinary
build, with the packages above and Go installed.

## OpenBSD

Not tried: a check of the build ran out of disk. Gio gives it X11 and EGL;
the tray falls back to `tray_other.go` (SNI is built for Linux and FreeBSD
only), and the `deviceinfo` file above stands in the way as on FreeBSD.

## Haiku

Read from sources in October 2026, nothing built. Unlike FreeBSD, a port
is not cheap: Gio has no backend for Haiku, and the graphics stack it
draws with elsewhere is not there.

Haiku is not a Unix by descent, but it has a good POSIX layer (libc,
threads, sockets, `mmap`), which is why Go and ffmpeg run there. What is
its own is the GUI: the `app_server` and the Be API in C++, with no X11 or
Wayland server.

### Go

- [go-haiku](https://github.com/Quad4-Software/go-haiku), a fork of Go
  for `haiku/amd64` and `haiku/386`, built on Haiku or cross-built from
  amd64 Haiku. Its source is Go 1.27.1, the version `go.mod` needs, but
  the latest release is `go1.26.8-haiku1`: 1.27 would be built from the
  source.
- cgo is on, which Gio needs. In this fork the `unix` build tag takes in
  Haiku, so the files built for `unix` or `!windows` here are built there.
- `golang.org/x/sys/unix` knows no Haiku upstream; the fork has its files
  only in the toolchain's own copy,
  [`src/cmd/vendor/golang.org/x/sys/unix`](https://github.com/Quad4-Software/go-haiku/tree/golang-1.26-haiku/src/cmd/vendor/golang.org/x/sys/unix).
  `internal/messenger/account/lock_unix.go` and the SQLite VFS import it:
  a fork of `x/sys` with those files would go in a `replace`.

### Graphics

Gio draws on Linux and the BSDs with OpenGL ES 3 through EGL, in an X11 or
Wayland window. On Haiku:

1. **A Haiku backend in the Gio fork**, as Gio has for Windows and macOS:
   windows, input, the clipboard and input methods on the Be API
   (`BWindow`, `BView`, `BMessage`), in C++ behind a C layer for cgo; and
   [Mesa's EGL platform for Haiku](https://gitlab.freedesktop.org/mesa/mesa/-/blob/main/src/egl/drivers/haiku/egl_haiku.cpp)
   to draw with, which gives OpenGL ES 3 and takes a `BitmapHook` as its
   native window, an object Mesa hands each drawn frame to. The surest way,
   and the largest: Gio's backends for Windows and macOS are over a
   thousand lines each.
2. **X512's Wayland compatibility layer**: a `libwayland-client.so`
   compatible in API and ABI, loaded into each program, with no compositor
   ([forum thread](https://discuss.haiku-os.org/t/my-progress-in-wayland-compatibility-layer/12373));
   GTK 3 and 4 run on it. Gio's Wayland backend could take it with `haiku`
   in its build tags, but GTK draws into shared memory buffers in software,
   while Gio needs EGL on Wayland, of which the thread says only that Mesa's
   EGL was built with Wayland. It hangs on others' experimental work.
3. **[Xlibe](https://github.com/waddlesplash/xlibe)**, Xlib on the Haiku
   API without an X server, does not fit: Gio's X11 backend needs xcb
   (`libX11-xcb`, `xkbcommon-x11`) and EGL on X11, and Xlibe has Xlib only.

There is little 3D acceleration: the drivers for AMD and Intel set modes
only, without 2D or 3D acceleration, and an accelerated driver for NVIDIA
Turing and Ampere came out as an alpha in January 2026
([OSnews](https://www.osnews.com/story/144097/haiku-gets-accelerated-nvidia-graphics-driver/),
[forum](https://discuss.haiku-os.org/t/whats-the-end-game-when-it-comes-to-3d-acceleration/18068)).
Gio would draw in software through Mesa there, and animations and the
blur would load the processor.

### The rest

| What | On Haiku |
|---|---|
| SQLite, the sandbox of decoders (wasm on `wazero`) | `wazero` has no compiler for Haiku, only its interpreter (`compilerPlatformSupports`: Linux, macOS, FreeBSD, NetBSD, Windows and a few others): the history cache, search and the H.264 and AAC decoders would be many times slower; video likely unusable without patching `wazero` |
| SQLite's file locks | `go-sqlite3` knows no Haiku and goes without locks; its portable `sqlite3_dotlk` tag is there for such systems |
| Sound (`oto`) | no backend for Haiku's Media Kit; the ALSA file is built for Linux, FreeBSD and NetBSD only, leaving PulseAudio's. Output through the Media Kit would be ours to write |
| Tray | Haiku has the Deskbar and its replicants; SNI over D-Bus does not apply |
| TPM | none |
| ffmpeg (the external player, voice recording) | ported to Haiku; its input devices for a microphone not looked into |

### What a port would take

A new backend for windows and graphics in the Gio fork, or a bet on the
experimental Wayland layer; a fork of `x/sys` with Haiku; a patched
`wazero`, or its interpreter; sound through the Media Kit; and an
unofficial Go toolchain. A project of its own, of weeks, drawing in
software. Were it started, the first thing to find out is whether EGL on
the Wayland layer works on Haiku: if it does, the backend comes down to
build tags and fixes.
