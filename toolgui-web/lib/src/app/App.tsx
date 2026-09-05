import React, { Component } from 'react'

import { Forest } from './Nodes'
import { clearState } from '../components/state'
import { AppConf } from './AppConf';
import { AppNavbar } from './AppNavbar';
import { AppBody } from './AppBody';
import { setIcon } from '../util/seticon';
import { AppError, Error } from './AppError';
import { UploadFunc } from './Upload';

const NOTIFY_TYPE_CREATE = 1
const NOTIFY_TYPE_UPDATE = 2
const NOTIFY_TYPE_DELETE = 3


interface AppProps {
  appConf: AppConf
  update: (event: any) => void
  upload: UploadFunc
}

interface AppState {
  forest: Forest
  running: boolean
  pageFound: boolean
  pageName: string
  error: Error | null
  darkMode: string
}

export class App extends Component<AppProps, AppState> {
  constructor(props: AppProps) {
    super(props);

    var pageName = ''
    if (this.props.appConf.hash_page_name_mode) {
      if (window.location.hash) {
        // should be #/{name}
        pageName = window.location.hash.substring(2)
      } else if (this.props.appConf.page_names.length > 0) {
        pageName = this.props.appConf.page_names[0]
      }
    } else {
      pageName = window.location.pathname.substring(1)
    }

    const curconf = this.props.appConf.page_confs[pageName]
    let pageFound = true
    if (curconf) {
      document.title = curconf.title
      if (curconf.emoji) {
        setIcon(curconf.emoji)
      }
    } else {
      document.title = 'Page not found'
      setIcon('❓')
      pageFound = false
    }

    this.state = {
      forest: new Forest([
        props.appConf.main_container_id,
        props.appConf.sidebar_container_id,
      ]),
      running: false,
      pageFound: pageFound,
      pageName: pageName,
      error: null,
      darkMode: localStorage.darkMode,
    }
  }

  startUpdate() {
    this.setState((prevState) => {
      const newForest = prevState.forest.swallowCopy()
      newForest.setToRemoving()

      return {
        running: true,
        forest: newForest,
        error: null,
      }
    })
  }

  receiveNotifyPack(pack: any) {
    switch (pack.type) {
      case NOTIFY_TYPE_CREATE: {
        this.setState((prevState) => {
          const newForest = prevState.forest.swallowCopy()
          newForest.createNode(pack.component, pack.container_id)

          return {
            forest: newForest,
          }
        })
        break
      }
      case NOTIFY_TYPE_UPDATE: {
        this.setState((prevState) => {
          const newForest = prevState.forest.swallowCopy()
          newForest.updateNode(pack.component)
          return {
            forest: newForest,
          }
        })
        break
      }
      case NOTIFY_TYPE_DELETE: {
        this.setState((prevState) => {
          const newForest = prevState.forest.swallowCopy()
          newForest.removeNode(pack.component_id)

          return {
            forest: newForest,
          }
        })
        break
      }
      default: {
        console.error('Notify pack type error', pack.type)
      }
    }
  }

  finishUpdate(pack: any) {
    this.setState((prevState) => {
      const newForest = prevState.forest.swallowCopy()
      newForest.removeNodeWithRemovingTag()
      var err: Error | null = null
      if (!pack.success) {
        err = {
          msg: pack.error
        }
      }

      return {
        running: false,
        forest: newForest,
        error: err,
      }
    })
  }

  clearState() {
    clearState()
  }

  render() {
    return (
      <div className="px-2">
        <AppNavbar
          appConf={this.props.appConf}
          running={this.state.running}
          pageFound={this.state.pageFound}
          pageName={this.state.pageName}
          rerun={() => { this.props.update({}) }}
          onChange={(darkMode) => {
            this.setState({
              darkMode: darkMode
            })
          }} />

        <AppBody
          appConf={this.props.appConf}
          pageFound={this.state.pageFound}
          forest={this.state.forest}
          update={(e) => { this.props.update(e) }}
          upload={async (f) => await this.props.upload(f)}
          darkMode={this.state.darkMode} />

        <AppError error={this.state.error} />
      </div >
    )
  }
}
