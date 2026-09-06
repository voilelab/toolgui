# Side Nav

The side nav is the left column of the app.

![side nav](sidenav.png)

It holds three parts, top to bottom:

1. The page list: one link per page in the App. The link of the current page is
   highlighted, and a page's emoji is shown in front of its title.
2. The page's [Sidebar container](../components/layout/container.md), when the
   page func puts anything in it.
3. The app controls:
   * Rerun: Rerun the Page Func without changing any state.
   * Dark/Light Mode Switch: Switch the theme of the app.
   * A spinner, shown while the app is running the Page Func.

On a narrow screen the column collapses behind a `Menu` button.
