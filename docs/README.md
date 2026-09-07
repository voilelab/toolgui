# ToolGUI docs

The [mdBook](https://rust-lang.github.io/mdBook/) source for the ToolGUI book.

Lives in this repository so an API change and its documentation land in the
same pull request.

## Layout

`src/` is the book. `design/` holds design notes and surveys that decide an API
before it exists; mdBook does not build them.

## Build

```shell
cargo install mdbook   # or grab a binary from the mdBook releases
mdbook serve           # http://localhost:3000, reloads on save
mdbook build           # static site, into book/
```

## Publishing

`.github/workflows/docs.yml` publishes to GitHub Pages from `main`, which the
Release workflow points at the latest release. So the published book describes
the released API, not `dev`; read `docs/src` on GitHub for the unreleased
state.
