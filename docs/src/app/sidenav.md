# Side Nav

The side nav is the left column of the app.

![side nav](sidenav.png)

It holds four parts, top to bottom:

1. The page list: one link per page in the App. The link of the current page is
   highlighted, and a page's emoji is shown in front of its title.
2. The page's [Sidebar container](../components/layout/container.md), when the
   page func puts anything in it.
3. The app controls:
   * Rerun: Rerun the Page Func without changing any state.
   * Dark/Light Mode Switch: Switch the theme of the app. Until it is used,
     the app follows the browser's `prefers-color-scheme`; the choice is
     remembered afterwards.
   * A spinner, shown while the app is running the Page Func.
4. The toolgui version the app was built against.

On a narrow screen the column collapses behind a `Menu` button.

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

