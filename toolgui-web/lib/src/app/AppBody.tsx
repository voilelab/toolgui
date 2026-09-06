import React, { Component, ReactNode } from "react";
import { AppConf } from "./AppConf";
import { TComponent } from "../components/factory";
import { MessagePageNotFound } from "./MessagePageNotFound";
import { UpdateEvent } from "./UpdateEvent";
import { Forest } from "./Nodes";
import { UploadFunc } from "./Upload";
import { ThemeMode } from "../util/theme";

interface AppBodyProps {
  appConf: AppConf
  pageFound: boolean
  forest: Forest
  update: (e: UpdateEvent) => void
  upload: UploadFunc
  themeMode: ThemeMode
}

// AppBody renders the page's main container. The page's sidebar container
// lives in AppSideNav, sharing the left column with the page list.
export class AppBody extends Component<AppBodyProps> {
  constructor(props: AppBodyProps) {
    super(props)
  }

  rootNode() {
    return this.props.forest.nodes[this.props.appConf.main_container_id]
  }

  render(): ReactNode {
    return (
      <div className="container">
        {this.props.pageFound ?
          <TComponent node={this.rootNode()}
            update={(e) => { this.props.update(e) }}
            upload={async (f) => await this.props.upload(f)}
            theme={this.props.themeMode} />
          : <MessagePageNotFound />}
      </div>
    )
  }
}
