// toolgui-wasm -offline marks index.html with this meta, naming the service
// worker it wrote next to it.
const META = 'meta[name="toolgui-sw"]'

// setupOffline registers the service worker of an -offline build, or removes
// the one an earlier -offline build left at this url.
export async function setupOffline() {
  if (!('serviceWorker' in navigator)) {
    return
  }

  const script = document.querySelector<HTMLMetaElement>(META)?.content
  if (script) {
    await navigator.serviceWorker.register(script)
    return
  }

  // Ours only: another app's worker may cover this page from a parent scope.
  const ours = new URL('sw.js', document.baseURI).href
  for (const reg of await navigator.serviceWorker.getRegistrations()) {
    const worker = reg.active || reg.waiting || reg.installing
    if (worker?.scriptURL !== ours) {
      continue
    }

    await reg.unregister()

    // The prefix sw.js names its caches with.
    const prefix = `toolgui-sw:${reg.scope}:`
    for (const name of await caches.keys()) {
      if (name.startsWith(prefix)) {
        await caches.delete(name)
      }
    }
  }
}
