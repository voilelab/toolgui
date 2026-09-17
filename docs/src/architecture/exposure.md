# Serving It Safely

ToolGUI has no login, no tokens and no permission model. Whoever reaches the
app runs it, with everything the tool itself can reach: the files it opens,
the databases it queries, the commands it shells out to. Treat "can open the
page" as "can do anything the tool does", and decide who that should be.

## Bind to loopback

`StartService(":3000")` listens on every interface, so the tool is on the
network the moment the machine is. The examples in these docs bind to
`127.0.0.1` instead, which keeps it on the machine it runs on:

```go
e := tgexec.NewWebExecutor(app)
e.StartService("127.0.0.1:3000")
```

That is the right default for a tool you run for yourself. It is also the
only thing standing between the tool and the rest of the network, so widen it
on purpose, not by copying a `:3000` out of a README.

## Origin check

The update websocket carries every event and every render of the app, and the
same-origin policy doesn't cover websockets: without a check, any page open in
the same browser could connect to `ws://localhost:3000/api/update/...`, press
the app's buttons and read back what they render.

So the handshake takes only an `Origin` whose host matches the one the request
asked for, and answers anything else with 403. Nothing has to be configured for
this; it is how the socket behaves.

## Limits

Nothing authenticates a connection, so what one can ask for is capped. The
service holds 1024 states at most, one per open page, and a connection that
finds no room is refused rather than handed one anyway. One upload is 1 GiB at
most, and it is stored under a component the page actually drew, so a caller
cannot keep a file per name it invents. An event is held to the same rule: it
writes the state under an id the last finished run drew, so a client cannot
fill a session with keys no component owns and no run would ever read or
release. One message on the update socket is 1 MiB at most, and a form event
nests 32 levels at most: every level of a form has its subtree read again, so
the two together are what stop one message from buying far more work than it
took to send. `StartService` also puts deadlines on sending a request's
headers, on sitting idle between requests, and on naming a state once the
websocket handshake is done.

A download goes the other way and is capped by nothing, because there is
nothing to cap: `DownloadFile` serves a file one of the page's own runs offered,
named by a token that is unguessable, that is looked up in the state which
offered it and nowhere else, and that is only accepted alongside the state id of
the connection asking. A token is therefore a bearer of nothing on its own, and
neither it nor the state id is ever put in a URL, where a link, a log line or a
`Referer` would carry it further than the fetch that needs it.

The caps are worth lowering on anything reachable by more than the person
running it:

```go
e.SetMaxStateCount(64)
e.SetMaxUploadSize(16 * 1024 * 1024)
e.SetMaxMessageSize(64 * 1024)
```

Raise the message cap instead for an app whose iframe or plugin components send
values of their own that are larger than 1 MiB. A message over the cap is
refused by its header, without being read into memory, and the page keeps its
connection.

These bound what one visitor costs. They are not a substitute for deciding who
reaches the page.

## Behind a reverse proxy

To serve the tool to more than the local machine, put it behind a proxy that
authenticates — an identity-aware proxy, an SSO gateway, whatever the team
already runs — and let the proxy reach ToolGUI over loopback.

A proxy that passes the browser's `Host` through needs nothing else. One that
rewrites it has to name the public origin, or the handshake will refuse the
browser it is proxying for:

```go
e.SetAllowedOrigins([]string{"https://tools.example.com"})
```

Each entry is a full origin, scheme and all, the way a browser sends it. The
app's own origin is always allowed on top of these.
