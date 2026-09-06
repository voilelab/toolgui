# Session Cache

Session-level cache stores data for one user's connection to the app.

ToolGUI has no separate API for it: the `State` **is** the session cache.
A session holds exactly one `State`, so everything on
[State Cache](state-cache.md) applies here too. What is worth knowing is how
long a session lives, which is up to the executor.

## Web executor

The executor keeps a pool of states, each with a `state_id`:

1. The client opens the update websocket and sends the `state_id` it has, or an
   empty one on the first connection.
2. The server hands back a new `state_id` when the client has none, or when the
   one it sent is gone. The client drops the components it drew and starts over
   on the fresh state.
3. The state is marked alive while the socket is open. On disconnect it stays in
   the pool, so a reconnect resumes on the same state.

A state that is no longer alive expires 5 minutes after its last use. The pool
is swept when a new state is created, so an expired state may outlive the
timeout on an idle server.

The `state_id` lives only in the page's JavaScript. A reload, or navigating to
another page, always starts a new session with an empty state.

## Desktop executor

One window is one session. `Start(pageName)` closes the previous session and
runs the new page on an empty state, the same as loading another page in the
browser.

## What does not belong here

Data that should outlive a session — anything shared by all users, or expensive
enough to compute once per process — belongs in an [App Cache](app-cache.md).
