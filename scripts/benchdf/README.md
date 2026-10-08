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
