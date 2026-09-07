# App

The App includes the info of

1. [Side Nav](sidenav.md): The left column of the app.
2. [Pages](page.md): The App holds an ordered map from page name to page's data.

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
