# ToolGUI Web

`lib` is the component library (Babel + `tsc --emitDeclarationOnly`).
`app` is the Vite bundle that the Go server embeds.

## Generate static file for web server

```shell
cd lib
yarn
yarn build
yarn typecheck
cd ../app
yarn
yarn build
```

`app` builds to `app/build/`, with hashed assets under `app/build/static/`.
Those paths are embedded by `../web.go` and served by `toolgui/tgexec`, and
`scripts/stub-assets.sh` creates the same layout as a placeholder — so
`build.outDir` / `build.assetsDir` in `app/vite.config.js` cannot change on
their own.

## Development

```shell
cd app
yarn start      # Vite dev server on :3000
yarn test       # Vitest + jsdom
yarn typecheck  # tsc --noEmit
```

## Dependency security

`yarn audit` from this directory reports **0 vulnerabilities**.

Transitive advisories are pinned via `resolutions` in the root
`package.json` — currently only `prismjs`, which `react-syntax-highlighter`
pulls in at a vulnerable major through `refractor`. `@types/react` is pinned
there too, to keep `@types/react-syntax-highlighter` from dragging in a
second major and breaking `yarn typecheck`.
