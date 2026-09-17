import { useEffect, useRef, useSyncExternalStore } from "react"

// The overlays on screen, oldest first, dialogs and popovers together.
//
// Neither kind can lean on Mantine for ESC. A Modal's closeOnEscape is a
// window handler per modal with no notion of a stack, so one press closes them
// all; a Popover's is a React handler on the dropdown, so it only fires while
// focus is already inside the panel, which it is not when the trigger opened
// it. One list settles both: the overlay opened last takes the key, whatever
// kind it is, so a popover opened inside a dialog closes before the dialog
// does.
//
// The list also settles which dialog is painted on top, since every Mantine
// Modal takes the same z-index and leaves the order to the DOM -- which
// follows where the page writes a dialog rather than when it opened.

export type OverlayKind = "dialog" | "popover"

interface Overlay {
  id: string
  kind: OverlayKind
}

const overlays: Overlay[] = []
const listeners = new Set<() => void>()

function announce() {
  listeners.forEach(notify => notify())
}

function subscribeToOverlays(notify: () => void) {
  listeners.add(notify)
  return () => { listeners.delete(notify) }
}

function openedOverlay(id: string, kind: OverlayKind) {
  overlays.push({ id, kind })
  announce()
}

function closedOverlay(id: string, kind: OverlayKind) {
  for (let at = overlays.length - 1; at >= 0; at--) {
    if (overlays[at].id === id && overlays[at].kind === kind) {
      overlays.splice(at, 1)
      announce()
      return
    }
  }
}

// depthOf counts the open overlays of the same kind below this one, -1 when it
// is not open itself. Only its own kind counts: a popover opening must not
// lift the dialog it sits in.
function depthOf(id: string, kind: OverlayKind): number {
  let depth = 0
  for (const overlay of overlays) {
    if (overlay.id === id && overlay.kind === kind) {
      return depth
    }

    if (overlay.kind === kind) {
      depth++
    }
  }

  return -1
}

function isTopmost(id: string, kind: OverlayKind): boolean {
  const top = overlays[overlays.length - 1]
  return top !== undefined && top.id === id && top.kind === kind
}

// useOverlay puts an open overlay on the stack and says where it sits: how
// many of its own kind are below it, and whether it is the topmost overlay of
// any kind, which is what answers ESC.
export function useOverlay(id: string, kind: OverlayKind, open: boolean) {
  useEffect(() => {
    if (!open) {
      return
    }

    openedOverlay(id, kind)
    return () => { closedOverlay(id, kind) }
  }, [open, id, kind])

  return {
    depth: useSyncExternalStore(
      subscribeToOverlays, () => depthOf(id, kind)),
    isTop: useSyncExternalStore(
      subscribeToOverlays, () => isTopmost(id, kind)),
  }
}

// Mantine components that handle ESC themselves, a Select with its dropdown
// open among them, mark the event this way. Same check Mantine's own modal
// makes, so a dropdown inside an overlay still closes on its own first.
function handledInside(target: EventTarget | null): boolean {
  return target instanceof Element &&
    target.getAttribute("data-mantine-stop-propagation") === "true"
}

// useEscapeToClose calls close on ESC while active. Pass the overlay's own
// isTop in: an overlay that is not on top must not take the key, and one that
// is on top but cannot be dismissed must swallow it rather than let the
// overlay under it have it, which is what leaving every other listener off
// gives.
export function useEscapeToClose(active: boolean, close: () => void) {
  // Read through a ref so that closing stays out of the dependencies below: a
  // fresh callback every render would tear the listener down and put it back
  // for nothing.
  const closeRef = useRef(close)
  closeRef.current = close

  useEffect(() => {
    if (!active) {
      return
    }

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key !== "Escape" || event.isComposing) {
        return
      }

      if (handledInside(event.target)) {
        return
      }

      closeRef.current()
    }

    window.addEventListener("keydown", onKeyDown, true)
    return () => { window.removeEventListener("keydown", onKeyDown, true) }
  }, [active])
}
