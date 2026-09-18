# Side Nav

The side nav is the left column of the app.

![side nav](sidenav.png)

It holds four parts, top to bottom:

1. The page list: one link per page in the App. The link of the current page is
   highlighted, and a page's emoji is shown in front of its title. A page the
   App declares [Hidden](page.md) is left off the list, except while it is the
   page being read.
2. The page's [Sidebar container](../components/layout/container.md), when the
   page func puts anything in it.
3. The app controls:
   * Rerun: Rerun the Page Func without changing any state.
   * Dark/Light Mode Switch: Switch the theme of the app. Until it is used,
     the app follows the browser's `prefers-color-scheme`; the choice is
     remembered afterwards.
   * A spinner, shown while the app is running the Page Func.
4. The toolgui version the app was built against.

The page list is the part that scrolls: an app with more pages than the
column is tall gets a scrollbar on the list, and the sidebar, the controls
and the version line under it keep their place.

On a narrow screen the column collapses behind a `Menu` button. What the
button opens is a bar that fits the screen: the list is capped there too, so
the controls under it are where they can be reached rather than below every
link the app has.

## Width

Drag the column's right edge to resize it, between 180px and 480px. The handle
is also a keyboard control: focus it and the arrow keys move the edge in 16px
steps, `Home` and `End` go to the bounds, and `Enter` — or a double click —
puts it back to the default 240px.

The button at the top of the column collapses it to just that button, handing
the whole width to the page; the page's Sidebar container keeps its state while
hidden. Both the width and the collapsed state are remembered in the browser,
so they survive moving between pages.

Neither applies on a narrow screen, where the column is a bar across the top
and the `Menu` button owns it.

## Version

The version line reads the module version out of the binary's build info, so an
app that depends on a released toolgui shows that tag with nothing to configure.
A build off an untagged checkout shows the pseudo-version the toolchain derives
from the commit, `v0.0.0-{date}-{revision}`.

Only a build whose info names no commit at all falls back to a version recorded
in the source: `go run`, a tree with no VCS metadata, or `-buildvcs=false`,
which the wails CLI passes on every build. A release records its own tag there;
anything built off `dev` reports `v0.0.0-unknown` rather than claiming a release
it is not.

Hide it with:

```go
app.SetShowVersion(false)
```

