# ToolGUI docs

The [mdBook](https://rust-lang.github.io/mdBook/) source for the ToolGUI book.

Lives in this repository so an API change and its documentation land in the
same pull request.

## Build

```shell
cargo install mdbook mdbook-mermaid   # or grab binaries from their releases
mdbook-mermaid install .              # once, writes the gitignored mermaid runtime
mdbook serve                          # http://localhost:3000, reloads on save
mdbook build                          # static site, into book/
```

`mdbook-mermaid` renders the ```` ```mermaid ```` blocks. `install` writes the
mermaid runtime that `book.toml` points at; it is gitignored rather than
committed, so run it once after cloning or `mdbook build` fails on the missing
files. Re-run it after upgrading `mdbook-mermaid` to refresh them.

## Publishing

`.github/workflows/docs.yml` publishes to GitHub Pages from `main`, which the
Release workflow points at the latest release. So the published book describes
the released API, not `dev`; read `docs/src` on GitHub for the unreleased
state.
