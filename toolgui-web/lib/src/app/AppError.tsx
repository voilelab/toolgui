import React, { Component, ReactNode } from "react";
import { Alert } from "@mantine/core";

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
      <div className="toolgui-page" style={{ paddingTop: '10px' }}>
        <Alert color="red">
          {this.props.error.msg}
        </Alert>
      </div>
    )
  }
}
