import React, { Component, ReactNode } from "react";
import { AppConf } from "./AppConf";
import { TComponent } from "../components/factory";
import { MessagePageNotFound } from "./MessagePageNotFound";
import { UpdateEvent } from "./UpdateEvent";
import { Forest } from "./Nodes";
import { UploadFunc } from "./Upload"
import { DownloadFunc } from "./Download";
import { ThemeMode } from "../util/theme";
import { PAGE_BOTTOM_ID } from "./PageBottom";

interface AppBodyProps {
  appConf: AppConf
  pageFound: boolean
  forest: Forest
  update: (e: UpdateEvent) => void
  upload: UploadFunc
  download: DownloadFunc
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
      <div className="toolgui-page">
        {this.props.pageFound ?
          <TComponent node={this.rootNode()}
            update={(e) => { this.props.update(e) }}
            upload={async (f, id) => await this.props.upload(f, id)}
            download={async (token) => await this.props.download(token)}
            theme={this.props.themeMode} />
          : <MessagePageNotFound />}
        {/* Where a pinned chat input is portaled to: last in the page, and
            sticky, so it stays at the bottom without covering the content. */}
        <div id={PAGE_BOTTOM_ID} className="toolgui-page-bottom" />
      </div>
    )
  }
}
