# benchdf

Measures what a DataFrame costs on a rerun that leaves its rows unchanged
(TG-94). Each page holds a Toggle and a DataFrame of 0 / 1000 / 3000 / 10000
rows (6 columns, typed cells, row keys, multi selection); flipping the Toggle
reruns the page and resends the whole table.

## Run

```shell
# Go side: pack size and marshal time, native and wasm
go test ./scripts/benchdf -run TestReport -v -args -report
go build -o /tmp/wasmtest ./scripts/wasmtest
GOOS=js GOARCH=wasm go test -exec /tmp/wasmtest ./scripts/benchdf -run TestReport -v \
  -args -report

# Browser side: click -> pack -> paint. Needs the web assets built and
# Playwright with its Chromium: npm i -g playwright && npx playwright install chromium
go run ./scripts/benchdf -addr 127.0.0.1:3100
go run ./cmd/toolgui-wasm serve -addr 127.0.0.1:3200 ./scripts/benchdf
NODE_PATH=$(npm root -g) node scripts/benchdf/measure.cjs
```

## Results (2026-10, headless Chromium, localhost)

Go side, per rerun:

| rows  | pack bytes | DataFrame share | marshal (native) | marshal (wasm) |
|------:|-----------:|----------------:|-----------------:|---------------:|
| 0     | 414        | 0%              | 0.002 ms         | 0.07 ms        |
| 1000  | 190 KB     | 99.8%           | 2.1 ms           | 17 ms          |
| 3000  | 569 KB     | 99.9%           | 7.9 ms           | 37 ms          |
| 10000 | 1.9 MB     | 100%            | 23.5 ms          | 104 ms         |

Browser, Toggle click to next paint (median of 30), JSON.parse included:

| rows  | web      | wasm     | web JSON.parse |
|------:|---------:|---------:|---------------:|
| 0     | 13 ms    | 14 ms    | 0              |
| 1000  | 46 ms    | 46 ms    | 1.1 ms         |
| 3000  | 64 ms    | 80 ms    | 2.4 ms         |
| 10000 | 97 ms    | 214 ms   | 7.1 ms         |

## TG-99: keep packs

The Go report now drives a `tgframe.Session`, so a rerun gets the keep pack
a real client gets for an unchanged component. Same machine, `main` before
and after (`first` is the first run, which always sends everything):

Go side, per rerun:

| rows  | first     | rerun before | rerun after | run (wasm) before | after  |
|------:|----------:|-------------:|------------:|------------------:|-------:|
| 0     | 414       | 414          | 230         | 0.2 ms            | 0.2 ms |
| 1000  | 190 KB    | 190 KB       | 345         | 17 ms             | 17 ms  |
| 3000  | 569 KB    | 569 KB       | 345         | 49 ms             | 50 ms  |
| 10000 | 1.9 MB    | 1.9 MB       | 345         | 188 ms            | 174 ms |

The session still marshals each component to compare it, so the Go run time
stays; what goes is the transfer, the client's parse and the recompute.

Browser, Toggle click to next paint (median of 30, two runs each):

| rows  | web before | web after | wasm before | wasm after |
|------:|-----------:|----------:|------------:|-----------:|
| 0     | 13.6/13.1  | 14.1/13.6 | 13.9/14.7   | 14.3/14.0  |
| 1000  | 49.1/49.5  | 42.6/44.9 | 47.1/46.6   | 45.7/46.0  |
| 3000  | 65.0/79.0  | 42.3/45.8 | 80.6/96.6   | 79.4/80.2  |
| 10000 | 97.3/97.0  | 64.8/79.7 | 230.2/247.1 | 180.9/211.6|
