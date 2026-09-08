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
