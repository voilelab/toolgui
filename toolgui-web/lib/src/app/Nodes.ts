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
}

export class Forest {
  nodes: { [key: string]: Node }
  rootNodeIDs: string[]
  runID: number

  constructor(rootNodeIDs: string[]) {
    this.rootNodeIDs = rootNodeIDs
    this.nodes = {}
    this.runID = NO_RUN_ID

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
    // does not inherit the old one's children.
    const node = oldNode && oldNode.props.name === props.name
      ? oldNode
      : new Node(key, props)

    node.props = props
    node.runID = this.runID
    node.parentKey = parentKey
    this.nodes[key] = node

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
    parentNode.children[index] = node
  }

  updateNode(key: string, props: any) {
    if (!(key in this.nodes)) {
      console.error('Try to update a node that doesn\'t exist:', key)
      return
    }

    this.nodes[key].props = props
  }

  removeNode(key: string) {
    const node = this.nodes[key]
    if (!node) {
      console.error('Try to remove a node that doesn\'t exist:', key)
      return
    }

    const parentNode = this.nodes[node.parentKey]
    if (parentNode) {
      parentNode.children = parentNode.children.filter(n => n !== node)
    }
    delete this.nodes[key]
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
    for (const node of Object.values(this.nodes)) {
      node.children = node.children.filter(n => n && n.runID === this.runID)
    }
  }
}
