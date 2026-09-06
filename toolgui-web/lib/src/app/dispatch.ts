import { App } from './App'

// dispatchPack route a pack from the backend to the matching App handler.
// Shared by every transport so they only have to deliver raw packs.
export function dispatchPack(app: App, pack: any) {
  if (pack.success !== undefined) {
    if (!pack.success) {
      console.error(pack)
    }

    app.finishUpdate(pack)
    return
  }

  if (pack.ready !== undefined) {
    if (!pack.ready) {
      console.error(pack)
    }

    app.startUpdate()
    return
  }

  app.receiveNotifyPack(pack)
}
