// Bundled rather than loaded from a CDN: a toolgui app has to work on a
// machine with no route to the internet. Declared here so the browser and
// desktop builds cannot ship different versions.
//
// Mantine first, then the app's own tokens and stylesheets, which read
// Mantine's variables and win where the two meet.
import '@mantine/core/styles.css'
import '@mantine/dates/styles.css'
import '@toolgui-web/lib/src/assets/css/theme.css'

import React, { Component } from 'react'
import { MantineProvider } from '@mantine/core'

import { Forest } from './Nodes'
import { clearState } from '../components/state'
import { AppConf } from './AppConf';
import { AppSideNav } from './AppSideNav';
import { AppBody } from './AppBody';
import { setIcon } from '../util/seticon';
import { emojize } from '../util/emoji';
import { AppError, Error } from './AppError';
import { UploadFunc } from './Upload';
import { ThemeMode, preferredThemeMode, themeModeManager } from '../util/theme';
import { ThemeModeSync } from './ThemeModeSync';

// pageNameFromLocation reads the page name off the URL.
function pageNameFromLocation(appConf: AppConf): string {
  if (!appConf.hash_page_name_mode) {
    return window.location.pathname.substring(1)
  }

  if (window.location.hash) {
    // should be #/{name}
    return window.location.hash.substring(2)
  }

  if (appConf.page_names.length > 0) {
    return appConf.page_names[0]
  }

  return ''
}

// documentTitle puts the app title after the page's, so a tab says which page
// of which app it holds. Either one alone stands on its own.
function documentTitle(appConf: AppConf, pageTitle: string): string {
  if (!appConf.title) {
    return pageTitle
  }

  if (!pageTitle) {
    return appConf.title
  }

  return `${pageTitle} - ${appConf.title}`
}

const NOTIFY_TYPE_CREATE = 1
const NOTIFY_TYPE_UPDATE = 2
const NOTIFY_TYPE_DELETE = 3


interface AppProps {
  appConf: AppConf
  update: (event: any) => void
  upload: UploadFunc

  // pageName and onNavigate let a transport without a URL (a desktop
  // webview) drive the routing. Left out, the page comes from window.location
  // and navigating moves the browser.
  pageName?: string
  onNavigate?: (name: string) => void
}

interface AppState {
  forest: Forest
  running: boolean
  pageFound: boolean
  pageName: string
  error: Error | null
}

export class App extends Component<AppProps, AppState> {
  // What to start in until the visitor has stored a choice of their own; the
  // stored one comes back through themeModeManager.
  private defaultColorScheme: ThemeMode

  constructor(props: AppProps) {
    super(props);

    const pageName = props.pageName !== undefined ?
      props.pageName : pageNameFromLocation(props.appConf)

    const curconf = this.props.appConf.page_confs[pageName]
    let pageFound = true
    if (curconf) {
      document.title = documentTitle(props.appConf, curconf.title)
      if (curconf.emoji) {
        setIcon(emojize(curconf.emoji))
      }
    } else {
      document.title = documentTitle(props.appConf, 'Page not found')
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
    }

    this.defaultColorScheme = preferredThemeMode()
  }

  startUpdate() {
    this.setState((prevState) => {
      const newForest = prevState.forest.swallowCopy()
      newForest.beginRun()

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
          newForest.createNode(pack.key, pack.parent_key, pack.index, pack.component)

          return {
            forest: newForest,
          }
        })
        break
      }
      case NOTIFY_TYPE_UPDATE: {
        this.setState((prevState) => {
          const newForest = prevState.forest.swallowCopy()
          newForest.updateNode(pack.key, pack.component)
          return {
            forest: newForest,
          }
        })
        break
      }
      case NOTIFY_TYPE_DELETE: {
        this.setState((prevState) => {
          const newForest = prevState.forest.swallowCopy()
          newForest.removeNode(pack.key)

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
      newForest.endRun(pack.success)
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
    // Mantine holds the color scheme, so there is one source of truth for the
    // theme; ThemeModeSync reads it back out for the app.
    return (
      <MantineProvider defaultColorScheme={this.defaultColorScheme}
        colorSchemeManager={themeModeManager}>
        <ThemeModeSync>
          {(themeMode) =>
            <div className="toolgui-shell">
              <AppSideNav
                appConf={this.props.appConf}
                forest={this.state.forest}
                running={this.state.running}
                pageFound={this.state.pageFound}
                pageName={this.state.pageName}
                onNavigate={this.props.onNavigate}
                rerun={() => { this.props.update({}) }}
                update={(e) => { this.props.update(e) }}
                upload={async (f, id) => await this.props.upload(f, id)}
                themeMode={themeMode} />

              <main className="toolgui-main">
                <AppBody
                  appConf={this.props.appConf}
                  pageFound={this.state.pageFound}
                  forest={this.state.forest}
                  update={(e) => { this.props.update(e) }}
                  upload={async (f, id) => await this.props.upload(f, id)}
                  themeMode={themeMode} />

                <AppError error={this.state.error} />
              </main>
            </div>}
        </ThemeModeSync>
      </MantineProvider>
    )
  }
}
