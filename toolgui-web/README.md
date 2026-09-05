# ToolGUI Web

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

## Dependency security

`yarn audit` from this directory. Transitive advisories are pinned via
`resolutions` in the root `package.json` — nested keys (`**/parent/child`)
target one consumer where several majors of a package coexist.

These stay open; they are build/test-time only (never shipped to the browser)
and have no fix reachable from `react-scripts@5.0.1`:

| Package | Why unfixed |
| --- | --- |
| `ajv@6` | Fix only in `ajv@8`; `schema-utils@2/3` and `eslint@8` need the v6 API. |
| `svgo@1` | Fix only in `svgo@2`; `@svgr/plugin-svgo@5` targets the v1 API. |
| `webpack-dev-server@4` | Fix only in v5; `react-scripts@5` is tied to the v4 API. |
| `uuid@8` | Fix only in v11. Used by `sockjs` for `v4()`, which the advisory (`buf` in v3/v5/v6) does not cover. |
| `@tootallnate/once@1` | Fix only in v2; pulled by `jsdom` via `http-proxy-agent`. |

Dropping `react-scripts` is the real fix for all five.
