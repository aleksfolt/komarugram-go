<div align="center">
<img src="./assets/logo_round.png" width="140" align="center" alt="KomaruGram">

# KomaruGram Go

Telegram Desktop клиент, написанный с нуля на языке Go. Обладает богатым функционалом и повышенным уровнем безопасности. Легко собирается и распространяется свободно на условиях минимально ограниченной лицензии.

[ [English](README.md) | [Русский] ]

[ [Прочее](./docs/README_FULL.md) | [Виджеты и UI](./docs/UI_COMPONENTS.md) ]

</div>

## Почему этот проект существует?

У официального клиента Telegram Desktop и его форков есть ряд недостатков: сильная зависимость от C++ кодовой базы и библиотек ([TDLib](https://github.com/tdlib/td), Qt), затратный по времени и ресурсам компьютера процесс сборки, а также высокий порог входа для разработки собственных модификаций. Кроме того, они используют copyleft лицензию, тогда как наш репозиторий полностью свободен в плане условий распространения.

## Что внутри?

За UI отвечает [Gio](https://gioui.org/) — кроссплатформенная библиотека, работающая по принципу immediate mode GUI. Выбор пал на нее, поскольку она позволяет избежать типичных ограничений фреймворков, не имеет зависимостей, не нарушает кроссплатформенную сборку, полностью совместима с Wayland и легко переносится на другие платформы.

[gotd/td](https://github.com/gotd/td) реализует MTProto и некоторые дополнительные возможности, служащие основой для любого взаимодействия с протоколом Telegram.

[wazero](https://github.com/wazero/wazero) используется в качестве высокопроизводительной WebAssembly песочницы для отрисовки стикеров. Это обеспечивает кроссплатформенность и гарантирует изоляцию кода без запуска отдельного процесса.

FFmpeg, mpv, VLC и Chromium доступны в качестве внешних интеграций, которые необходимо предоставить самостоятельно. Однако приложение будет работать и без них, если пользователя это устраивает.

Без FFmpeg или если это выбрано в настройках, GIF и анимированные аватары воспроизводятся H.264-декодером FFmpeg, скомпилированным в WebAssembly. Он не встроен в приложение: он загружается при первой необходимости из [libavcodec-wasm](https://github.com/komarugif/libavcodec-wasm), где лежат его исходники и скрипт сборки. Чтобы использовать свою сборку, укажите в переменной окружения `KOMARUGRAM_AVCDEC` путь к файлу `avcdec.wasm` или ссылку на него.

Голосовые сообщения и музыка воспроизводятся в самом приложении, без FFmpeg и внешнего плеера, прямо во время загрузки: Opus — через libopus, MP3, FLAC и WAV — через dr_libs, все скомпилированы в WebAssembly и встроены. AAC (M4A) воспроизводится декодером Fraunhofer FDK AAC, который загружается при первой необходимости из [fdk-aac-wasm](https://github.com/komarugif/fdk-aac-wasm), где лежат его исходники; `KOMARUGRAM_AACDEC` указывает на свою сборку.

## С чего начать

Создайте копию этого репозитория. Убедитесь, что на вашем компьютере установлен **[Go 1.27.1](https://go.dev/dl/)** — это минимально необходимая версия. В Linux могут потребоваться библиотеки разработки Wayland и X11.

Если вам не нужно создавать исполняемый файл, вы можете использовать эту команду для запуска программы в различных операционных системах.
```
go run ./cmd/messenger
```

Если же вам необходимо создать исполняемый файл со значком, выполните следующие действия.

### **Linux:**
**Просто сделайте go build:**
```
GOOS=linux go build -ldflags="-s -w" -buildvcs=false -o komarugram ./cmd/messenger
```

### **Windows:**

**Установите Gio cmd тулзы**:
```
go install gioui.org/cmd/gogio@latest
```
**Сборка приложения с иконкой:**
```
gogio -ldflags="-s -w" -icon=./assets/logo_round.png -target=windows -o komarugram.exe ./cmd/messenger
```

Процесс сборки может потребовать до 4 ГБ оперативной памяти. Учитывайте это и закрывайте ненужные приложения во время первой сборки. Все последующие сборки должны выполняться практически мгновенно.

## Переносимость

KomaruGram Go может быть портирован на множество операционных систем благодаря своей архитектуре и минимальному количеству зависимостей. Загляните в [PLATFORMS.md](./docs/PLATFORMS.md), если вам интересно изучить особенности под каждую из платформ.

| OS | Состояние |
|---|---|
| Windows 10/11 | ✅ Первоклассная поддержка |
| Linux (Wayland) | ✅ Первоклассная поддержка |
| Linux (X11) | ✅ Поддерживается, насколько этого позволяет X11 |
| MacOS | ✅ Первоклассная поддержка |
| FreeBSD | ⚠️ Должен собираться, не проверено вживую |
| OpenBSD | ⚠️ В теории должен собираться, не проверено |
| NetBSD | ❓ Возможен перенос, с некоторыми оговорками |
| Android | ❓ Возможен перенос, с некоторыми оговорками |
| Windows 7 | ❓ Возможен перенос, с некоторыми оговорками |
| Haiku OS | ❓ Возможен перенос, с многочисленными оговорками |


## Вайбкодинг

Код в этом проекте был написан преимущественно языковыми моделями GPT-6 Astra и Claude Opus 5.5. За концепцию, выбор стека технологий и библиотек, а также за контроль качества отвечает сопровождающий проекта.

## Благодарности

Мы благодарим создателя [gotd/td](https://github.com/gotd/td) за отличную библиотеку, а также [Gio](https://gioui.org/) — за архитектуру, которая одновременно проста и масштабируема для больших приложений. Брендинг предоставлен бесплатно проектом [t.me/komarugram](https://t.me/komarugram). Спасибо augustwise и SvatoshGPT за предоставленные подписки на Claude Code и Codex под нужды этого проекта.

## Вам может быть интересно

Если вы искали тот KomaruGram, который форк Telegram Desktop, [вам сюда](https://github.com/svatoshgpt/komarugram). Наш проект преследует схожие цели, но написан с нуля на Go и не является форком Telegram Desktop

## Лицензия

При работе с этим проектом у вас нет никаких лицензионных обязательств в отношении изменения или распространения кода. Некоторые зависимости требуют соблюдения авторских прав, но не накладывают ограничений на сам код (используются разрешительные лицензии, совместимые с MIT).

Исключений два. Оба вынесены из этого репозитория и из бинарника, загружаются во время работы, и их легко заменить или не использовать:

- H.264-декодер — из FFmpeg, по лицензии LGPL 2.1 или более поздней версии. Он лежит в [libavcodec-wasm](https://github.com/komarugif/libavcodec-wasm); `KOMARUGRAM_AVCDEC` заменяет его.
- AAC-декодер — Fraunhofer FDK AAC Codec Library, по собственной лицензии Fraunhofer: свободно копируется и распространяется вместе с исходниками, но не даёт патентных прав. AAC защищён патентами, лицензии на которые выдаёт [Via LA](https://www.via-la.com): для использования декодера, особенно коммерческого, такая лицензия может понадобиться. Поэтому Debian считает его несвободным. Он лежит в [fdk-aac-wasm](https://github.com/komarugif/fdk-aac-wasm); `KOMARUGRAM_AACDEC` заменяет его, а без него не воспроизводится только звук в M4A.

## Лицензии сторонних компонентов

Модули Go, которые входят в `messenger` и `kitchen`, по списку `go version -m` для сборок под Linux и Windows:

| Компонент | Лицензия |
|---|---|
| Go (стандартная библиотека и рантайм) | BSD-3-Clause |
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
| [ncruces/go-sqlite3-wasm](https://github.com/ncruces/go-sqlite3-wasm) | MIT-0; SQLite внутри — общественное достояние |
| [lukechampine.com/adiantum](https://github.com/lukechampine/adiantum) | MIT |
| [go-text/typesetting](https://github.com/go-text/typesetting) | Unlicense OR BSD-3-Clause |
| [dlclark/regexp2](https://github.com/dlclark/regexp2) | MIT |
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
| [Шрифты Go](https://go.dev/blog/go-fonts) (`gioui.org/font/gofont`) | BSD-3-Clause |

Встроенные и загружаемые файлы:

| Компонент | Лицензия |
|---|---|
| `pkg/vp9/vpxdec.wasm`: [libvpx](https://github.com/webmproject/libvpx) с [SIMDe](https://github.com/simd-everywhere/simde) | BSD-3-Clause (libvpx), MIT (SIMDe) |
| `pkg/opus/opusdec.wasm`: [libopus](https://github.com/xiph/opus) (уведомление — в `pkg/opus/COPYING.libopus`) | BSD-3-Clause |
| `pkg/drdec/drdec.wasm`: [dr_libs](https://github.com/mackron/dr_libs) (dr_mp3, dr_flac, dr_wav; `pkg/drdec/LICENSE.dr_libs`) | Unlicense OR MIT-0 |
| `pkg/lottie/tlottie.wasm`: [tlottie](https://github.com/dkaraush/tlottie) и стандартная библиотека Rust | MIT; MIT OR Apache-2.0 |
| WASI libc и compiler-rt внутри модулей выше ([wasi-sdk](https://github.com/WebAssembly/wasi-sdk)) | MIT, Apache-2.0 WITH LLVM-exception |
| `pkg/prism/grammars.dat.gz`: грамматики [Prism.js](https://prismjs.com) 1.29.0 вместе с `pkg/prism`, портом токенизатора [libprisma](https://github.com/desktop-app/libprisma) (`pkg/prism/LICENSE.prism`) | MIT |
| `pkg/miniapp/assets/telegram-web-app.js`: SDK Mini Apps от Telegram | © Telegram, лицензия не указана |
| `avcdec.wasm`, загружается во время работы: H.264-декодер FFmpeg ([libavcodec-wasm](https://github.com/komarugif/libavcodec-wasm)) | LGPL-2.1-or-later |
| `aacdec.wasm`, загружается во время работы: декодер Fraunhofer FDK AAC ([fdk-aac-wasm](https://github.com/komarugif/fdk-aac-wasm)) | Лицензия Fraunhofer FDK AAC, без патентных прав |

Полные тексты лежат в файле `LICENSE` каждого модуля (`go env GOMODCACHE`) и в репозиториях по ссылкам выше.
