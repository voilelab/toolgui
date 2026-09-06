import React, { Component, ReactNode } from "react";

export interface Error {
  msg: string
}

interface AppErrorProps {
  error: Error | null
}

export class AppError extends Component<AppErrorProps> {
  render(): ReactNode {
    if (!this.props.error) {
      return <></>
    }

    return (
      <div className="container" style={{ paddingTop: '10px' }}>
        <article className="message is-danger">
          <div className="message-body">
            {this.props.error.msg}
          </div>
        </article>
      </div>
    )
  }
}