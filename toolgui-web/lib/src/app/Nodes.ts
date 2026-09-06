// NO_RUN_ID is the run id of a node no run has claimed yet.
const NO_RUN_ID = 0

export class Node {
  props: any
  children: Node[]
  // The run that last sent this node. Anything older is stale at endRun.
  runID: number
  parentID: string

  constructor(props: any) {
    this.props = props
    this.children = []
    this.runID = NO_RUN_ID
    this.parentID = ''
  }
}

export class Forest {
  nodes: { [id: string]: Node }
  rootNodeIDs: string[]
  runID: number

  constructor(rootNodeIDs: string[]) {
    this.rootNodeIDs = rootNodeIDs
    this.nodes = {}
    this.runID = NO_RUN_ID

    for (const id of rootNodeIDs) {
      this.nodes[id] = new Node({
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

  createNode(props: any, parentID: string) {
    const nodeID: string = props.id
    const oldNode = this.nodes[nodeID]

    if (oldNode && oldNode.runID === this.runID) {
      console.error('Duplicated component id:', nodeID)
      return
    }

    // Detach from the parent of the previous run. The parent can be gone when
    // a delete pack removed it without its children.
    if (oldNode) {
      const prevParent = this.nodes[oldNode.parentID]
      if (prevParent) {
        const idx = prevParent.children.indexOf(oldNode)
        if (idx != -1) {
          prevParent.children.splice(idx, 1)
        }
      }
    }

    // create or modify node in node pool
    if (oldNode) {
      oldNode.props = props
    } else {
      this.nodes[nodeID] = new Node(props)
    }

    this.nodes[nodeID].runID = this.runID
    this.nodes[nodeID].parentID = parentID

    // Insert before the first sibling this run has not sent yet, so the
    // children keep the order the page function wrote them in.
    const parentNode = this.nodes[parentID]
    var idx = 0
    for (var i = 0; i < parentNode.children.length; i++) {
      const prevNode = parentNode.children[i]
      if (prevNode.runID !== this.runID) {
        break
      }
      idx = i + 1
    }
    parentNode.children.splice(idx, 0, this.nodes[nodeID])
  }

  updateNode(props: any) {
    const compID: string = props.id

    if (!(compID in this.nodes)) {
      console.error('Try to update a node that doesn\'t exist:', compID)
      return
    }

    this.nodes[compID].props = props
  }

  removeNode(nodeID: string) {
    if (!(nodeID in this.nodes)) {
      console.error('Try to remove a node that doesn\'t exist:', nodeID)
      return
    }

    const parentID = this.nodes[nodeID].parentID
    const idx = this.nodes[parentID].children.findIndex(n => n.props.id === nodeID)
    this.nodes[parentID].children.splice(idx, 1)
    delete this.nodes[nodeID]
  }

  // endRun closes a run and drops every node it did not send. A run that
  // failed leaves the tree alone: the page function stopped partway through,
  // so what it did not send is missing rather than gone.
  endRun(success: boolean) {
    if (!success) {
      return
    }

    for (const [id, node] of Object.entries(this.nodes)) {
      if (node.runID !== this.runID) {
        delete this.nodes[id]
      }
    }

    for (const node of Object.values(this.nodes)) {
      node.children = node.children.filter(n => n.runID === this.runID)
    }
  }
}
