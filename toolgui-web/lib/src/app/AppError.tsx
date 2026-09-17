import React, { Component, ReactNode } from "react";
import { Alert } from "@mantine/core";

export interface Error {
  msg: string

  // id names the server log line holding what really went wrong, for an error
  // the server only tells the browser the kind of. Undefined for a message
  // that is already the whole of it.
  id?: string
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
      <div className="toolgui-page" style={{ paddingTop: '10px' }}>
        <Alert color="red">
          {this.props.error.msg}
          {this.props.error.id &&
            <div style={{ fontSize: '0.85em', opacity: 0.7, marginTop: '4px' }}>
              error id: {this.props.error.id}
            </div>}
        </Alert>
      </div>
    )
  }
}
