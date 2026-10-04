# App

The App includes the info of

1. [Side Nav](sidenav.md): The left column of the app.
2. [Pages](page.md): The App holds an ordered map from page name to page's data.
3. [Menu](menu.md): A menubar along the top, for an App that declares one.

![layout](layout.png)

## Title

The App can name itself:

```go
app.SetTitle("My Tool")
```

The browser tab then reads `{page title} - {app title}`, the title names the
app in the [web manifest](manifest.md), and it titles the desktop window
unless [`Conf`](../hello-world/desktop.md#conf) sets its own. An App with no
title leaves the tab to the page title alone.

## Icon

By default the browser tab shows the current page's emoji. An App can set
one icon for every page instead:

```go
app.SetIcon("assets/favicon.svg")
```

The url is whatever the page can load: one under
[`SetAssets`](manifest.md), or the `-assets` of a
[browser build](../hello-world/wasm.md#icon). It takes effect once the
frontend has the app's config; the icon an installed app shows comes from the
[web manifest](manifest.md) `"icons"` instead.
