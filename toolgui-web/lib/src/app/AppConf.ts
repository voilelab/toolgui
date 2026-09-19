// MenuNode is one node of the app's menu tree. The tree is declared on the Go
// App rather than by a page func, so it is the same for every run.
export type MenuNode = MenuTextNode | MenuSeparatorNode | MenuSubmenuNode

// MenuTextNode is an item that sends a click. Its id already carries the
// reserved `menu_item_` prefix the Go side reads it back under, so it cannot
// be the id of a component on the page.
export interface MenuTextNode {
  type: 'text'
  label: string
  id: string

  // The key combination the item also fires on, normalized by the Go side
  // ("CmdOrCtrl+Shift+O"), and absent for an item that declared none. The
  // desktop lets the OS dispatch it; here the menubar listens for it. See
  // accelerator.ts.
  accelerator?: string
}

export interface MenuSeparatorNode {
  type: 'separator'
}

// MenuSubmenuNode holds items of its own. children is absent for an empty one.
export interface MenuSubmenuNode {
  type: 'submenu'
  label: string
  children?: MenuNode[]
}

// PageConf is one page, as the Go App declares it.
export interface PageConf {
  name: string
  title: string
  emoji: string

  // Whether the side nav leaves the page off its list. Absent for a page the
  // list holds, which is most of them.
  hidden?: boolean
}

export interface AppConf {
  page_names: string[]
  page_confs: { [page_name: string]: PageConf }

  title: string,

  main_container_id: string,
  sidebar_container_id: string,

  hash_page_name_mode: boolean,

  version: string,
  show_version: boolean,

  // Absent for an app that declares no menu, which is what keeps the menubar
  // row out of the DOM.
  menu?: MenuNode[],
}
