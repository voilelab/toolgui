export class Node {
  props: any
  children: Node[]
  removing: boolean
  parentID: string

  constructor(props: any) {
    this.props = props
    this.children = []
    this.removing = false
    this.parentID = ''
  }
}

export class Forest {
  nodes: { [id: string]: Node }
  rootNodeIDs: string[]

  constructor(rootNodeIDs: string[]) {
    this.rootNodeIDs = rootNodeIDs
    this.nodes = {}

    for (const id of rootNodeIDs) {
      this.nodes[id] = new Node({
        name: 'container_component',
        id: id,
      })
    }
  }

  swallowCopy(): Forest {
    const ret = new Forest(this.rootNodeIDs)
    ret.nodes = { ...this.nodes }
    return ret
  }

  setToRemoving() {
    for (const nodeID in this.nodes) {
      if (nodeID in this.rootNodeIDs) {
        continue
      }

      this.nodes[nodeID].removing = true
    }
  }

  createNode(props: any, parentID: string) {
    const nodeID: string = props.id

    if (nodeID in this.nodes && !this.nodes[nodeID].removing) {
      console.error('Depulicated component id:', nodeID)
      return
    }

    // remove node from old parent
    if (nodeID in this.nodes) {
      const parentID = this.nodes[nodeID].parentID
      const idx = this.nodes[parentID].children.findIndex(n => n.props.id === nodeID)
      // TBD: Why we need check here?
      if (idx != -1) {
        this.nodes[parentID].children.splice(idx, 1)
      }
    }

    // create or modify node in node pool
    const oldNode = this.nodes[nodeID]
    if (oldNode) {
      oldNode.props = props
      oldNode.removing = false
    } else {
      this.nodes[nodeID] = new Node(props)
    }

    this.nodes[nodeID].parentID = parentID

    // find first that first removing=true and insert at that index
    const parentNode = this.nodes[parentID]
    var idx = 0
    for (var i = 0; i < parentNode.children.length; i++) {
      const prevNode = parentNode.children[i]
      if (prevNode.removing) {
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

  removeNodeWithRemovingTag() {
    const removingID = new Set<string>()
    for (const [key, node] of Object.entries(this.nodes)) {
      if (node.removing) {
        removingID.add(key)
      }
    }

    for (const id in removingID) {
      delete this.nodes[id]
    }

    for (const [key, node] of Object.entries(this.nodes)) {
      node.children = node.children.filter((n) => {
        return !(removingID.has(n.props.id))
      })
    }
  }
}