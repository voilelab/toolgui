// Puts the live demo on a component page.
//
// A page marks the spot with `<div data-toolgui-demo="button">`, holding the
// screenshot it used to show on its own. This turns that into an iframe onto
// the demo app's page for the same component -- /demo/#/button, running the
// very examples the page includes its code from. Loaded from book.toml's
// additional-js, so it is the one copy of the embedding: a new component page
// writes the attribute and nothing else.
//
// What the page wrote is left alone where the demo cannot run, which is the
// fallback demo-embed.css writes the note for.

(function () {
  'use strict'

  // The demo app is published beside the book, under /demo. index.html rather
  // than the directory: a plain static file server need not serve one for the
  // other.
  var DEMO_PAGE = 'demo/index.html'

  // A fixed height, the way Streamlit embeds an app: the frame cannot ask the
  // app how tall it is, and one that comes up short scrolls. A demo that
  // needs more room says so with data-toolgui-demo-height.
  var DEFAULT_HEIGHT = 320

  // bookRoot is the path from this page up to the root of the book. mdBook
  // writes it for its own scripts, on every page.
  function bookRoot() {
    return typeof path_to_root === 'string' ? path_to_root : ''
  }

  // demoURL is the demo app's page for name. `embed` drops the app's own nav
  // and trims its padding, which is a display mode of the frontend rather
  // than something the app declares -- hence the query string. The page lives
  // in the hash: the site is static and cannot route paths.
  function demoURL(name, embed) {
    return bookRoot() + DEMO_PAGE + (embed ? '?embed' : '') + '#/' + name
  }

  // embed replaces the box's fallback with the running demo.
  function embed(box) {
    var name = box.getAttribute('data-toolgui-demo')
    if (!name) {
      return
    }

    var frame = document.createElement('iframe')
    frame.className = 'toolgui-demo-frame'
    frame.title = name + ' demo'
    frame.src = demoURL(name, true)
    frame.style.height = height(box) + 'px'

    // Not fetched until it is scrolled near: a reader who never reaches the
    // demo never downloads the wasm binary behind it.
    frame.loading = 'lazy'

    var link = document.createElement('a')
    link.className = 'toolgui-demo-link'
    link.href = demoURL(name, false)
    link.target = '_blank'
    link.rel = 'noopener'
    link.textContent = 'Open this demo in its own tab'

    // The screenshot was standing in for this; it goes.
    box.textContent = ''
    box.appendChild(frame)
    box.appendChild(link)
    box.setAttribute('data-toolgui-demo-state', 'live')
  }

  function height(box) {
    var given = parseInt(box.getAttribute('data-toolgui-demo-height'), 10)
    return given > 0 ? given : DEFAULT_HEIGHT
  }

  function run() {
    var boxes = document.querySelectorAll('[data-toolgui-demo]')
    if (boxes.length === 0) {
      return
    }

    // No wasm, no demo: the page keeps the screenshot and the note under it.
    if (typeof WebAssembly === 'undefined') {
      return
    }

    // mdBook's print page is every page at once, so it holds every marker in
    // the book. Screenshots are what belongs on paper anyway.
    if (/\/print\.html$/.test(window.location.pathname)) {
      return
    }

    // One frame to a page, so a page cannot start a second wasm instance --
    // each one is a worker with a copy of the binary in it. A component's
    // page in the demo app already shows every example it has, so there is
    // never a second demo to draw. cmd/toolgui-demo's tests hold the book to
    // that; this is what the rule means if one ever slips through.
    embed(boxes[0])

    for (var i = 1; i < boxes.length; i++) {
      console.warn('toolgui: one demo to a page, leaving',
        boxes[i].getAttribute('data-toolgui-demo'), 'as its screenshot')
    }
  }

  // additional-js is loaded at the end of the body, so the page is there --
  // unless a future mdBook moves it.
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', run)
  } else {
    run()
  }
})()
