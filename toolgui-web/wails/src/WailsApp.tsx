import React from "react"

import { SessionApp } from "@toolgui-web/lib"
import { connect } from "./api/backend"

// WailsApp has no address bar, so the page lives in memory.
export function WailsApp() {
  return <SessionApp connect={connect} />
}
