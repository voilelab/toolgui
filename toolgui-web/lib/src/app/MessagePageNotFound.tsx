import React from "react"

export function MessagePageNotFound() {
  return (
    <div className="columns is-centered">
      <div className="column is-three-quarters">
        <article className="message is-warning">
          <div className="message-header">
            <p>Oops! Page not found.</p>
          </div>
          <div className="message-body">
            We're sorry, the page you requested was not found.
            Try using the navigation menu to find what you're looking for.
          </div>
        </article>
      </div>
    </div>
  )
}