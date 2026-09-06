// GUEST_SCRIPT is prepended to the iframe's srcDoc, so it is the host that
// puts it there and no cross-origin access is needed to install it. Everything
// it does afterwards goes through postMessage, which keeps working once the
// sandbox drops allow-same-origin.
//
// Plain ES5-ish JS on purpose: it runs in the guest document, untouched by the
// build.
export const GUEST_SCRIPT = `
(function () {
  var VERSION = 1

  var renderListeners = []
  var uploads = {}
  var nextRequestID = 0

  function post(message) {
    message.toolgui = VERSION
    parent.postMessage(message, '*')
  }

  var toolgui = {
    // Latest values from the host. Undefined until the first render message.
    id: undefined,
    props: undefined,
    theme: undefined,

    // onRender runs fn on every render message, and once immediately if a
    // render has already arrived.
    onRender: function (fn) {
      renderListeners.push(fn)
      if (toolgui.props !== undefined) {
        fn(toolgui.props, toolgui.theme)
      }
    },

    // update sends a value back to the server. It lands in the state under
    // this iframe's own id, which the host fills in.
    update: function (value) {
      post({ type: 'update', value: value })
    },

    // upload sends a File to the server.
    upload: function (file) {
      return new Promise(function (resolve) {
        var requestID = String(++nextRequestID)
        uploads[requestID] = resolve
        post({ type: 'upload', requestID: requestID, file: file })
      })
    },

    // autoHeight reports the content height to the host as it changes.
    // Pair it with Height: "auto" on the Go side.
    autoHeight: function () {
      function report() {
        // Measure the body, not documentElement: documentElement is at least
        // as tall as the frame, so it would report the height the host just
        // set and never shrink. Body margins are outside its box.
        var style = window.getComputedStyle(document.body)
        var height = document.body.getBoundingClientRect().height +
          parseFloat(style.marginTop) + parseFloat(style.marginBottom)

        post({ type: 'resize', height: Math.ceil(height) })
      }

      if (window.ResizeObserver) {
        new ResizeObserver(report).observe(document.body)
      }

      window.addEventListener('load', report)
      report()
    },
  }

  window.addEventListener('message', function (event) {
    if (event.source !== parent) {
      return
    }

    var data = event.data
    if (!data || data.toolgui !== VERSION) {
      return
    }

    if (data.type === 'render') {
      toolgui.id = data.id
      toolgui.props = data.props
      toolgui.theme = data.theme

      for (var i = 0; i < renderListeners.length; i++) {
        renderListeners[i](toolgui.props, toolgui.theme)
      }

      return
    }

    if (data.type === 'upload_result') {
      var resolve = uploads[data.requestID]
      if (resolve) {
        delete uploads[data.requestID]
        resolve({ ok: data.ok, error: data.error })
      }
    }
  })

  window.toolgui = toolgui

  // The guest speaks first: postMessage does not queue, so the host cannot
  // safely send anything until it knows this listener exists.
  post({ type: 'ready' })
})()
`
