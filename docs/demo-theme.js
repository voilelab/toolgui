// Keeps the demo in the page's iframe on the same theme as the book.
//
// The book and the demo app are published as one site -- the book at the
// root, the app under /demo -- so they share an origin, and with it
// localStorage. The book keeps its theme under `mdbook-theme`; the app keeps
// its own under `theme_mode`, and follows a write to that key from outside
// the document through the storage event (toolgui-web/lib/src/util/theme.ts).
//
// So mirroring the one key into the other is the whole of the sync: a frame
// already on the page follows without being reloaded or talked to, and one
// that has not loaded yet reads the right theme when it does.
//
// Loaded from book.toml's additional-js, ahead of demo-embed.js, so the key
// is already right when the frame that reads it is built.

(function () {
  'use strict'

  // The theme the app is put in for each theme the book offers. The app knows
  // light and dark and nothing else, so the three dark books all map onto the
  // one dark app -- ayu is not a theme the demo can be in.
  var APP_THEME = {
    light: 'light',
    rust: 'light',
    coal: 'dark',
    navy: 'dark',
    ayu: 'dark',
  }

  // Where the app keeps its theme. mdBook keeps its own under
  // `mdbook-theme`, but the class on <html> is what is actually being shown
  // -- storage holds a choice the reader made, and says nothing until they
  // have made one -- so that is what this reads.
  var APP_KEY = 'theme_mode'

  // The last theme written. A class changes on <html> for the sidebar too, so
  // most of what arrives here is not a theme change at all; and writing only
  // a real change is what leaves a reader who picked a theme inside the frame
  // with the theme they picked.
  var written = null

  // bookTheme is the theme the book is being read in, or null where <html>
  // carries no theme this knows -- a future mdBook marking it some other way,
  // which is left alone rather than guessed at.
  function bookTheme() {
    var classes = document.documentElement.classList

    for (var name in APP_THEME) {
      if (classes.contains(name)) {
        return name
      }
    }

    return null
  }

  function sync() {
    var theme = bookTheme()
    if (!theme) {
      return
    }

    var mode = APP_THEME[theme]
    if (mode === written) {
      return
    }

    written = mode

    try {
      window.localStorage.setItem(APP_KEY, mode)
    } catch (e) {
      // Site data is blocked or full. The demo stays on its own theme, which
      // is a frame out of step with the book rather than a page that broke.
    }
  }

  function run() {
    // Only where there is a demo to theme: every other page of the book has
    // no frame, and the app's own preference is not the book's to write.
    if (!document.querySelector('[data-toolgui-demo]')) {
      return
    }

    sync()

    // The reader switching the book's theme is a class swapped on <html>,
    // which is mdBook's own doing and has no event on it.
    if (typeof MutationObserver === 'function') {
      new MutationObserver(sync).observe(document.documentElement, {
        attributes: true,
        attributeFilter: ['class'],
      })
    }
  }

  // additional-js is loaded at the end of the body, so the page is there --
  // unless a future mdBook moves it. Ahead of demo-embed.js either way: this
  // script is loaded first, so its listener is the first to run.
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', run)
  } else {
    run()
  }
})()
