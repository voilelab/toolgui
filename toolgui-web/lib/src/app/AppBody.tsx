import React, { Component, ReactNode } from "react";
import { AppConf } from "./AppConf";
import { TComponent } from "../components/factory";
import { MessagePageNotFound } from "./MessagePageNotFound";
import { UpdateEvent } from "./UpdateEvent";
import { Forest } from "./Nodes";
import { UploadFunc } from "./Upload";

interface AppBodyProps {
  appConf: AppConf
  pageFound: boolean
  forest: Forest
  update: (e: UpdateEvent) => void
  upload: UploadFunc
  darkMode: string
}

export class AppBody extends Component<AppBodyProps> {
  constructor(props: AppBodyProps) {
    super(props)
  }

  rootNode() {
    return this.props.forest.nodes[this.props.appConf.main_container_id]
  }

  sidebarNode() {
    return this.props.forest.nodes[this.props.appConf.sidebar_container_id]
  }

  render(): ReactNode {
    const hasSidebar = this.sidebarNode().children.length > 0
    return (
      <div className="container" style={{ paddingTop: '60px' }}>
        {this.props.pageFound ?
          <section className="columns is-fullheight">
            {hasSidebar ?
              <aside className="column is-3">
                <div style={{ position: 'sticky', overflow: 'auto', top: '60px' }}>
                  <TComponent node={this.sidebarNode()}
                    update={(e) => { this.props.update(e) }}
                    upload={async (f) => await this.props.upload(f)}
                    theme={this.props.darkMode} />
                </div>
              </aside> : ''}
            <div className={`column ${hasSidebar ? 'is-9' : 'is-12'}`}>
              <TComponent node={this.rootNode()}
                update={(e) => { this.props.update(e) }}
                upload={async (f) => await this.props.upload(f)}
                theme={this.props.darkMode} />
            </div>
          </section>
          : <MessagePageNotFound />}
      </div>
    )
  }
}