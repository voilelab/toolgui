// NO_RUN_ID is the run id of a node no run has claimed yet.
const NO_RUN_ID = 0

export class Node {
  props: any
  children: Node[]
  // The run that last sent this node. Anything older is stale at endRun.
  runID: number
  // Where the node sits, "<parent key>/<index>". Identity across runs.
  key: string
  parentKey: string

  constructor(key: string, props: any) {
    this.props = props
    this.children = []
    this.runID = NO_RUN_ID
    this.key = key
    this.parentKey = ''
  }

  // reactKey is what the renderer keys this node by. A component that names
  // itself keeps that name, so its React state follows it when a conditional
  // above it moves it. The rest are keyed by position and type: the renderer
  // reaches every component through one TComponent, so a key that ignored the
  // type would let React carry one component's hooks into another's.
  get reactKey(): string {
    return this.props.id || `${this.key}:${this.props.name}`
  }

  clone(): Node {
    const ret = new Node(this.key, this.props)
    ret.children = [...this.children]
    ret.runID = this.runID
    ret.parentKey = this.parentKey
    return ret
  }
}

// A Forest copy never changes a node's props or children in place: the node is
// replaced, and so is every node above it. The renderer skips a node whose
// reference has not changed, so everything a pack did not touch stays as is.
export class Forest {
  nodes: { [key: string]: Node }
  rootNodeIDs: string[]
  runID: number
  // Nodes this copy made, which no earlier copy shares.
  private owned: Set<Node>

  constructor(rootNodeIDs: string[]) {
    this.rootNodeIDs = rootNodeIDs
    this.nodes = {}
    this.runID = NO_RUN_ID
    this.owned = new Set()

    for (const id of rootNodeIDs) {
      this.nodes[id] = new Node(id, {
        name: 'container_component',
        id: id,
      })
    }
  }

  swallowCopy(): Forest {
    // Constructed with no roots on purpose: every node comes from the copy
    // below, and App copies the forest once per incoming pack.
    const ret = new Forest([])
    ret.rootNodeIDs = this.rootNodeIDs
    ret.nodes = { ...this.nodes }
    ret.runID = this.runID
    return ret
  }

  // own returns the node at key as one this copy may change, replacing it and
  // its ancestors with copies the first time.
  private own(key: string): Node {
    const node = this.nodes[key]
    if (this.owned.has(node)) {
      return node
    }

    const copy = node.clone()
    this.owned.add(copy)
    this.nodes[key] = copy

    const parentNode = this.nodes[node.parentKey]
    const index = parentNode ? parentNode.children.indexOf(node) : -1
    if (index >= 0) {
      this.own(node.parentKey).children[index] = copy
    }

    return copy
  }

  // beginRun opens a run. Every node the run sends is stamped with the new id,
  // so whatever still carries an older one at endRun was not re-sent.
  beginRun() {
    this.runID++

    // The server never re-sends the root containers, so the run claims them
    // here. Their survival is structural, not a special case in endRun.
    for (const id of this.rootNodeIDs) {
      this.nodes[id].runID = this.runID
    }
  }

  createNode(key: string, parentKey: string, index: number, props: any) {
    const parentNode = this.nodes[parentKey]
    if (!parentNode) {
      console.error('Component sent into a container that doesn\'t exist:', parentKey)
      return
    }

    const oldNode = this.nodes[key]

    // A different component type at this position is a different node, so it
    // does not inherit the old one's children. The same props keep the old
    // reference, so the renderer skips it.
    let node: Node
    if (oldNode && oldNode.props.name === props.name) {
      node = oldNode.props === props ? oldNode : this.own(key)
      node.props = props
    } else {
      node = new Node(key, props)
      this.owned.add(node)
      this.nodes[key] = node
    }

    // Neither is rendered, so a node shared with an earlier copy can take them.
    node.runID = this.runID
    node.parentKey = parentKey

    // A node this run has not sent renders under the same key when a named
    // component moves between positions. Two children under one key is what
    // this is all meant to avoid, and endRun is too late, so retire it now.
    for (const [staleKey, stale] of Object.entries(this.nodes)) {
      if (staleKey !== key && stale.runID !== this.runID
        && stale.reactKey === node.reactKey) {
        this.removeNode(staleKey)
      }
    }

    // The index is the component's position among its container's children,
    // counted by the container as the page function writes it.
    if (this.nodes[parentKey].children[index] !== node) {
      this.own(parentKey).children[index] = node
    }
  }

  // keepNode places the node already at key as a create of its own props
  // would: same props reference, so nothing below it recomputes.
  keepNode(key: string, parentKey: string, index: number) {
    const node = this.nodes[key]
    if (!node) {
      console.error('Try to keep a node that doesn\'t exist:', key)
      return
    }

    this.createNode(key, parentKey, index, node.props)
  }

  updateNode(key: string, props: any) {
    if (!(key in this.nodes)) {
      console.error('Try to update a node that doesn\'t exist:', key)
      return
    }

    if (this.nodes[key].props !== props) {
      this.own(key).props = props
    }
  }

  // removeNode takes the node at key off the tree, and the subtree under it
  // with it: a container that is gone does not leave its contents behind to be
  // inherited by whatever lands on its key next.
  //
  // A key the forest doesn't have is already in the state this asks for, which
  // is what lets a container clear a place only an earlier run wrote.
  removeNode(key: string) {
    const node = this.nodes[key]
    if (!node) {
      return
    }

    const parentNode = this.nodes[node.parentKey]
    if (parentNode && parentNode.children.includes(node)) {
      const owned = this.own(node.parentKey)
      owned.children = owned.children.filter(n => n !== node)
    }
    delete this.nodes[key]

    for (const child of [...node.children]) {
      if (child) {
        this.removeNode(child.key)
      }
    }
  }

  // endRun closes a run and drops every node it did not send. A run that
  // failed leaves the tree alone: the page function stopped partway through,
  // so what it did not send is missing rather than gone.
  endRun(success: boolean) {
    if (!success) {
      return
    }

    for (const [key, node] of Object.entries(this.nodes)) {
      if (node.runID !== this.runID) {
        delete this.nodes[key]
      }
    }

    // Also closes the gaps a shorter run left in the children arrays.
    // Array.from, not every: every skips the gaps.
    const live = (n: Node | undefined) => !!n && n.runID === this.runID
    for (const key of Object.keys(this.nodes)) {
      if (!Array.from(this.nodes[key].children).every(live)) {
        const node = this.own(key)
        node.children = node.children.filter(live)
      }
    }
  }
}
