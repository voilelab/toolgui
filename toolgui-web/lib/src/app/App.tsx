// Bundled rather than loaded from a CDN: a toolgui app has to work on a
// machine with no route to the internet. Declared here so the browser and
// desktop builds cannot ship different versions.
//
// Mantine first, then the app's own tokens and stylesheets, which read
// Mantine's variables and win where the two meet.
import '@mantine/core/styles.css'
import '@mantine/dates/styles.css'
import '@mantine/notifications/styles.css'
import '@toolgui-web/lib/src/assets/css/theme.css'

import React, { Component } from 'react'
import { MantineProvider, getDefaultZIndex } from '@mantine/core'
import { Notifications } from '@mantine/notifications'

import { Forest } from './Nodes'
import { clearState } from '../components/state'
import { AppConf } from './AppConf';
import { AppSideNav } from './AppSideNav';
import { AppMenuBar } from './AppMenuBar';
import { AppBody } from './AppBody';
import { setIcon, setIconURL } from '../util/seticon';
import { emojize } from '../util/emoji';
import { AppError, Error } from './AppError';
import { UploadFunc } from './Upload';
import { DownloadFunc } from './Download';
import { ThemeMode, preferredThemeMode, themeModeManager } from '../util/theme';
import { ThemeModeSync } from './ThemeModeSync';
import { PageQuery, encodeQuery, pageFromLocation, pageHref } from './pageurl';
import { PageNav, PageNavContext, newPageNav } from './PageNav';

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
const NOTIFY_TYPE_KEEP = 4


interface AppProps {
  appConf: AppConf
  update: (event: any) => void
  upload: UploadFunc
  download: DownloadFunc

  // pageName and onNavigate let a transport without a URL (a desktop
  // webview) drive the routing. Left out, the page comes from window.location
  // and navigating moves the browser.
  pageName?: string
  // query is the page query pageName was opened with, encoded.
  query?: string
  onNavigate?: (name: string, query: string) => void

  // onReplaceQuery takes the encoded page query Params.ReplaceQuery set. Left
  // out, it replaces the query in the address bar.
  onReplaceQuery?: (query: string) => void

  // embed drops the app's own chrome -- the page list, the controls, the
  // version line -- and leaves the page itself. For an iframe, where the
  // document around it is the navigation and the width is someone else's to
  // spend.
  embed?: boolean
}

interface AppState {
  forest: Forest
  running: boolean
  pageFound: boolean
  pageName: string
  error: Error | null

  // The page query as last opened or replaced, and the PageNav built on it:
  // a new PageNav redraws the links for the sticky keys it carries.
  query: string
  pageNav: PageNav
}

export class App extends Component<AppProps, AppState> {
  // What to start in until the visitor has stored a choice of their own; the
  // stored one comes back through themeModeManager.
  private defaultColorScheme: ThemeMode

  constructor(props: AppProps) {
    super(props);

    const loc = props.pageName !== undefined ?
      { name: props.pageName, query: props.query || '' } :
      pageFromLocation(window.location,
        props.appConf.hash_page_name_mode, props.appConf.page_names)
    const pageName = loc.name

    const curconf = this.props.appConf.page_confs[pageName]
    let pageFound = true
    if (curconf) {
      document.title = documentTitle(props.appConf, curconf.title)
    } else {
      document.title = documentTitle(props.appConf, 'Page not found')
      pageFound = false
    }

    // The app's icon wins over the page emoji.
    if (props.appConf.icon) {
      setIconURL(props.appConf.icon)
    } else if (!curconf) {
      setIcon('❓')
    } else if (curconf.emoji) {
      setIcon(emojize(curconf.emoji))
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
      query: loc.query,
      pageNav: this.newPageNav(loc.query),
    }

    this.defaultColorScheme = preferredThemeMode()
  }

  // newPageNav is how PageLink and the side nav move to another page.
  newPageNav(query: string): PageNav {
    return newPageNav(this.props.appConf.hash_page_name_mode,
      this.props.onNavigate, query, this.props.appConf.sticky_query || [])
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
      case NOTIFY_TYPE_KEEP: {
        this.setState((prevState) => {
          const newForest = prevState.forest.swallowCopy()
          newForest.keepNode(pack.key, pack.parent_key, pack.index)

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

  // replaceQuery swaps the page query for the one Params.ReplaceQuery set: no
  // reload, no new session, no history entry.
  replaceQuery(query: PageQuery) {
    const q = encodeQuery(query)
    if (q !== this.state.query) {
      this.setState({ query: q, pageNav: this.newPageNav(q) })
    }

    if (this.props.onReplaceQuery) {
      this.props.onReplaceQuery(q)
      return
    }

    window.history.replaceState(window.history.state, '',
      pageHref(this.state.pageName, q, this.props.appConf.hash_page_name_mode))
  }

  // navigate opens the page Params.Navigate named, as a PageLink click would:
  // a new session and a history entry. Only a page of this app is taken.
  navigate(name: string, query: PageQuery) {
    if (!this.props.appConf.page_confs[name]) {
      console.error('Navigate to unknown page', name)
      return
    }

    this.state.pageNav.navigate(name, query)
  }

  finishUpdate(pack: any) {
    this.setState((prevState) => {
      const newForest = prevState.forest.swallowCopy()
      newForest.endRun(pack.success)
      var err: Error | null = null
      if (!pack.success) {
        err = {
          msg: pack.error,
          id: pack.error_id,
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
    // Embed mode is the page on its own, so the menubar goes with the rest of
    // the app's chrome. An empty tree counts as none.
    const menu = !this.props.embed && this.props.appConf.menu?.length ?
      this.props.appConf.menu : null

    // Mantine holds the color scheme, so there is one source of truth for the
    // theme; ThemeModeSync reads it back out for the app.
    return (
      <MantineProvider defaultColorScheme={this.defaultColorScheme}
        colorSchemeManager={themeModeManager}>
        <PageNavContext.Provider value={this.state.pageNav}>
          {/* Where every Toast lands. Above the dialogs and their popovers: a
              toast is transient and says what just happened, so whatever it was
              fired from must not cover it. */}
          <Notifications zIndex={getDefaultZIndex('max')} />

          <ThemeModeSync>
            {(themeMode) =>
              // The frame is a column: the menubar row, and the shell's two
              // columns under it. An app that declares no menu gets no row --
              // the frame is then the shell in a wrapper, and --tg-menubar-h
              // stays 0, so every 100vh the shell is built on still holds.
              <div className={`toolgui-frame ${menu ? 'has-menubar' : ''}`}>
                {menu ?
                  <AppMenuBar menu={menu}
                    update={(e) => { this.props.update(e) }} /> : ''}

                <div className={`toolgui-shell ${this.props.embed ? 'is-embed' : ''}`}>
                  {this.props.embed ? '' :
                    <AppSideNav
                      appConf={this.props.appConf}
                      forest={this.state.forest}
                      running={this.state.running}
                      pageFound={this.state.pageFound}
                      pageName={this.state.pageName}
                      pageNav={this.state.pageNav}
                      rerun={() => { this.props.update({}) }}
                      update={(e) => { this.props.update(e) }}
                      upload={async (f, id) => await this.props.upload(f, id)}
                      download={async (token) => await this.props.download(token)}
                      themeMode={themeMode} />}

                  <main className="toolgui-main">
                    <AppBody
                      appConf={this.props.appConf}
                      pageFound={this.state.pageFound}
                      forest={this.state.forest}
                      update={(e) => { this.props.update(e) }}
                      upload={async (f, id) => await this.props.upload(f, id)}
                      download={async (token) => await this.props.download(token)}
                      themeMode={themeMode} />

                    <AppError error={this.state.error} />
                  </main>
                </div>
              </div>}
          </ThemeModeSync>
        </PageNavContext.Provider>
      </MantineProvider>
    )
  }
}
