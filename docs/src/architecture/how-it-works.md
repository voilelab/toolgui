# How it works?

## Basic

![ui-state-pagefunc](ui-state-pagefunc.png)

The key concept is that in the **Page Function**,
the UI component interact immediately with the running logic.

For example:

```go
if tgcomp.Button(p.State, p.Main, "Click me") {
    tgcomp.Text(p.Main, "Hi")
}
```

In the first call (For example: When the web page is entering.),
The `Click me` button will render, but the `Hi` will not render.
Since the button is not clicked in the first round.

When the user clicks the button, the **Page Function** will be called again.
In this round, `tgcomp.Button` will return `true` and `Hi` will render.

How?

Since we declare a button whose id is `Click me`, and when the button is clicked,
we store `true` with key `Click me` into `p.State`.

When entering `tgcomp.Button`, it will check if there is any data stored
in `p.State` with key `Click me`.

## Server-Client Architecture

<svg viewBox="0 0 980 300" role="img" aria-label="Server-client architecture" style="width:100%;height:auto;max-width:980px;font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace">
  <defs>
    <marker id="sa-arrow" viewBox="0 0 10 10" refX="9" refY="5"
            markerWidth="6" markerHeight="6" orient="auto-start-reverse">
      <path d="M0,0 L10,5 L0,10 z" fill="currentColor"/>
    </marker>
  </defs>
  <g fill="none" stroke="currentColor" stroke-width="1.5">
    <rect x="30" y="30" width="140" height="70" rx="4"/>
    <rect x="290" y="30" width="140" height="70" rx="4"/>
    <rect x="30" y="200" width="140" height="70" rx="4"/>
    <rect x="290" y="200" width="140" height="70" rx="4"/>
    <rect x="550" y="200" width="140" height="70" rx="4"/>
    <rect x="810" y="200" width="140" height="70" rx="4"/>
  </g>
  <g stroke="currentColor" stroke-width="1.5" marker-end="url(#sa-arrow)">
    <line x1="100" y1="102" x2="100" y2="198" marker-start="url(#sa-arrow)"/>
    <line x1="330" y1="198" x2="330" y2="102"/>
    <line x1="390" y1="102" x2="390" y2="198"/>
    <line x1="170" y1="220" x2="288" y2="220"/>
    <line x1="430" y1="220" x2="548" y2="220"/>
    <line x1="690" y1="220" x2="808" y2="220"/>
    <line x1="808" y1="250" x2="692" y2="250"/>
    <line x1="548" y1="250" x2="432" y2="250"/>
    <line x1="288" y1="250" x2="172" y2="250"/>
  </g>
  <g fill="currentColor" font-size="14" text-anchor="middle">
    <text x="100" y="70">State</text>
    <text x="360" y="70">State Storage</text>
    <text x="100" y="240">Client</text>
    <text x="360" y="240">Executor</text>
    <text x="620" y="240">Session</text>
    <text x="880" y="240">Page Func</text>
  </g>
  <g fill="currentColor" font-size="11">
    <text x="324" y="155" text-anchor="end">state ID</text>
    <text x="396" y="155">State</text>
    <g text-anchor="middle">
      <text x="229" y="214">event</text>
      <text x="489" y="214">event</text>
      <text x="749" y="214">State</text>
      <text x="749" y="266">notify</text>
      <text x="489" y="266">packs</text>
      <text x="229" y="266">packs</text>
    </g>
  </g>
</svg>

Server-Client need to handle a more complex part: multiple state for multiple users.
Hence we need a `state_id` for each state.

The executor owns the transport and the state pool — how long a state lives in
it is on [Session Cache](session-cache.md). Everything from the event onwards
is the Session's, and it is the same on the desktop.

`packs` is what travels back: a ready pack when a run starts, one notify pack
per component change, and a result pack when the run ends.

File upload is the one thing off this path. The client POSTs to `/api/files`
with its `state_id` and the handler writes the bytes straight into the state,
so it touches neither the socket nor the Session.

## Session

The piece that turns an event into a page run is `tgframe.Session`.
It holds a page name, a `State`, and one function to send packs to the client:

```go
func NewSession(app *App, pageName string, state *State, send SendPackFunc) (*Session, error)
```

That is all it needs, so it does not know what the transport is.
An executor only feeds it events:

```go
session.HandleRawEvent(bs)
```

A `Session` is safe for concurrent use and serializes its runs: a new event
interrupts the page func still running from the previous one, so the client
never receives two runs interleaved. The interrupted run panics with
`ErrUpdateInterrupt`, which the session recovers.

This is why the web executor and the
[desktop executor](../hello-world/desktop.md) share the whole app logic. They
differ only in how packs and events travel: a websocket and a `state_id` pool on
the web, bound methods on a single window on the desktop.
