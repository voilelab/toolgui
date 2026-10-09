// Service worker toolgui-wasm writes for -offline: it keeps a copy of the
// site's files, so an installed app opens with no network.
//
// Network first: online, the site is served as if there were no worker, so a
// new build, or one without -offline, takes effect on the next load. The copy
// is only for when the network fails, and is refreshed by every load that
// reaches it.
//
// toolgui-wasm fills in VERSION, a hash of the files, FILES, their paths
// relative to this script, and LAZY, the -lazy-assets paths: those are not
// fetched on install, only kept once the page fetches them.

const VERSION = '__VERSION__'
const FILES = __FILES__
const LAZY = __LAZY__

// The scope keeps two apps on one origin out of each other's caches.
const PREFIX = `toolgui-sw:${self.registration.scope}:`
const CACHE = PREFIX + VERSION

const URLS = new Set(FILES.map((name) => new URL(name, self.location).href))
const LAZY_URLS = new Set(LAZY.map((name) => new URL(name, self.location).href))
const INDEX = new URL('index.html', self.location).href

self.addEventListener('install', (event) => {
  event.waitUntil((async () => {
    const cache = await caches.open(CACHE)
    // no-cache revalidates, so a stale http cache entry of an older build
    // cannot end up in the copy.
    await cache.addAll([...URLS].map((url) => new Request(url, { cache: 'no-cache' })))
    await self.skipWaiting()
  })())
})

self.addEventListener('activate', (event) => {
  event.waitUntil((async () => {
    for (const name of await caches.keys()) {
      if (name.startsWith(PREFIX) && name !== CACHE) {
        await caches.delete(name)
      }
    }

    // So the first visit, which loaded before this worker, works offline
    // from here on too.
    await self.clients.claim()
  })())
})

self.addEventListener('fetch', (event) => {
  const req = event.request
  if (req.method !== 'GET') {
    return
  }

  // A navigation's url keeps its hash, and ?embed is read by the page.
  const url = new URL(req.url)
  url.search = ''
  url.hash = ''

  // The app root is index.html. Anything else not in FILES or LAZY is left
  // alone: the site may share its directory with other pages.
  let key = url.href
  if (req.mode === 'navigate' && key === self.registration.scope) {
    key = INDEX
  }
  if (!URLS.has(key) && !LAZY_URLS.has(key)) {
    return
  }

  event.respondWith((async () => {
    const cache = await caches.open(CACHE)

    try {
      const resp = await fetch(req)
      // Not awaited: the page streams app.wasm and shows how much arrived.
      if (resp.ok) {
        event.waitUntil(cache.put(key, resp.clone()))
      }
      return resp
    } catch (err) {
      const hit = await cache.match(key)
      if (hit) {
        return hit
      }
      throw err
    }
  })())
})
