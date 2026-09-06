# State Storage

ToolGUI stores data at three levels, longest lived first:

* [App Cache](app-cache.md): for the whole process, shared by every user.
  Not provided by ToolGUI — the developer implements it.
* [Session Cache](session-cache.md): for one user's connection to the app.
* [State Cache](state-cache.md): for the page currently shown, through
  `p.State`.

Why cache at all:

- Faster access: Frequently used data can be retrieved from the cache much faster than recalculating it or fetching it from an external source every time. This improves the application's overall performance.

- Reduced resource usage: By avoiding redundant calculations and external data fetching, the app can conserve resources like CPU and network bandwidth.
