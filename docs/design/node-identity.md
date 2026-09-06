# Node identity and lifecycle: how Streamlit does it

Survey for [issue #67](https://github.com/voilelab/toolgui/issues/67). It
describes Streamlit's model, compares it against what `toolgui` does today, and
lays out the options and their cost. It does not implement anything.

Streamlit source is quoted from `develop` as of September 2026; paths are given
by file and symbol rather than line, because that tree moves.

## The question

Both frameworks re-run a page function on every interaction and re-send every
component. The framework then has to answer: **is this component the one that
was here last run, or a new one?** The answer decides whether React updates the
props in place or remounts, whether widget state survives, and which nodes are
garbage.

`toolgui` answers it with a single string, the component id, derived from the
component's content (`tcutil.NormalID`, `HashedID`, `RandID`). Two consequences
were verified against the current `dev`:

* Two components with identical content produce identical ids, and
  `Forest.createNode` drops the second with a `console.error`
  ([`Nodes.ts:51`](../../toolgui-web/lib/src/app/Nodes.ts)). Six components sent,
  four rendered.
* `Divider` uses `RandID`, so it is a different component every run: its `<hr>`
  remounts on every rerun and the old id stays in `Forest.nodes` forever
  (measured: the pool grows by two per rerun on a two-divider page).

## Streamlit's model in one line

**Position identifies a node; content never does.** A second identifier, derived
from content, exists only for widget *state*, and a collision there is a raised
exception, not a dropped element.

The two are deliberately separate, and that separation is the part worth
copying.

### 1. Position: the delta path

Every `st.*` call writes to a cursor held on the script run context
(`lib/streamlit/cursor.py`). A `RunningCursor` holds `root_container`,
`parent_path` and an `index` that starts at 0 and increments on every element:

```python
def make_delta_path(root_container, parent_path, index):
    delta_path = [root_container]
    delta_path.extend(parent_path)
    delta_path.append(index)
    return delta_path
```

`lock_element()` reserves the current index and advances; `open_block()` makes a
child cursor whose `parent_path` is `(*parent_path, index)` and advances. So the
third element inside the second column of the main container is
`[0, 1, 1, 2]` — a per-container call ordinal, nested.

Two identical `st.write("hi")` calls are simply index 3 and index 4. Nothing
compares their content, so there is nothing to collide.

### 2. The path travels on the wire

`DeltaGenerator._enqueue` stamps it onto the message
(`lib/streamlit/delta_generator.py`):

```python
msg.metadata.delta_path[:] = dg._cursor.delta_path
```

This is the structural difference from `toolgui`. A `notifyPackCreate` carries
`container_id` and the component, but no position; the frontend reconstructs
order by inserting before the first child still tagged `removing`
([`Nodes.ts:86`](../../toolgui-web/lib/src/app/Nodes.ts)). Streamlit's frontend
never guesses — it is told the index.

### 3. The frontend holds a tree, not a pool

`AppRoot` (`frontend/lib/src/render-tree/AppRoot.ts`) holds one immutable
`BlockNode` root with `main`, `sidebar`, `event` and `bottom` beneath it.
`applyDelta` walks `deltaPath` and rebuilds the spine:

```ts
SetNodeByDeltaPathVisitor.setNodeAtPath(this.root, deltaPath, elementNode, scriptRunId)
```

`SetNodeByDeltaPathVisitor` consumes one index per level, copies the children
array at the target level, and returns new `BlockNode`s all the way up — so
React's shallow comparison re-renders exactly the changed subtree. Out-of-range
index is an error; one past the end is an append.

There is no `{[id]: Node}` map anywhere. That is why Streamlit has no equivalent
of the leak in `removeNodeWithRemovingTag`: a node that is not in the tree is
not anywhere.

`addBlock` has one extra rule worth noting — replacing a block of the same type
inherits the old block's children, "to preserve widget state and the React state
of all elements". Blocks are containers; their children are re-sent right after,
so this is what keeps a `st.container`'s contents from flickering.

### 4. Lifecycle: stamp with a run id, prune when the run ends

Every node records the `scriptRunId` it was created in (`AppNode.scriptRunId`,
default `NO_SCRIPT_RUN_ID`). Deltas overwrite nodes at their path during the
run; nothing is deleted while it runs. When `ScriptFinished` arrives with
`FINISHED_SUCCESSFULLY`, `App.tsx` calls:

```ts
elements: elements.clearStaleNodes(scriptRunId, fragmentIdsThisRun)
```

`ClearStaleNodeVisitor` then drops every node whose `scriptRunId` is not the
current one, recursively, and returns the same node object when nothing under it
changed. A shorter page this run means the extra tail nodes are simply stale.

Three details that map directly onto issue #67's second half:

* **A run id, not a boolean.** `toolgui`'s `removing` flag needs a matching
  clear on every touched node, and stale state if the clear is missed. A stamp
  compared against "the current run" needs no clearing.
* **Root containers cannot be pruned.** `clearStaleNodes` maps the root's
  children through `ensureBlockNode`, which substitutes a fresh empty
  `BlockNode` for anything the visitor removed. `toolgui` tries to express the
  same rule with `nodeID in this.rootNodeIDs` over an array
  ([`Nodes.ts:39`](../../toolgui-web/lib/src/app/Nodes.ts)), which is always
  false — the root containers are never re-sent by the Go side
  ([`app.go:17`](../../toolgui/tgframe/app.go)), so nothing ever clears their
  flag. Streamlit makes the root's survival structural instead of conditional.
* **Pruning is skipped when the run did not finish cleanly.** Only
  `FINISHED_SUCCESSFULLY` and the fragment equivalent prune;
  `FINISHED_EARLY_FOR_RERUN` deliberately does not, "to avoid flickering of
  elements where they disappear for a moment and then are readded".

Widget *state* is collected separately, and afterwards:
`AppRoot.getActiveIds()` walks the pruned tree for live element and block ids,
and `removeInactiveWidgetState()` drops the rest.

### 5. Content-derived ids exist — for state, and only for stateful elements

`st.markdown` and `st.text` never compute an id. Widgets call
`compute_and_register_element_id` (`lib/streamlit/elements/lib/utils.py`), which
hashes the element type, the user `key`, the command's own kwargs, the active
script hash, the form id and the root container, and formats the result as
`$$ID-<hash>-<user_key>`.

That id is a session-state key. It is not used to place the element in the tree
— placement is the delta path, computed independently. This is the split
`toolgui` does not have: `Button` compares `s.GetClickID() == comp.ID`
([`button.go:60`](../../toolgui/tgcomp/tcinput/button.go)) and `Textbox` reads
`s.GetString(comp.ID)`, using the same string that `Forest` uses as a tree key.

### 6. A duplicate id is a raised error

`_register_element_id` keeps two per-run sets and raises on collision:

```python
if user_key and not ctx.shared.widget_user_keys_this_run.check_and_add(user_key):
    raise StreamlitDuplicateElementKey(user_key)
if not ctx.shared.widget_ids_this_run.check_and_add(element_id):
    raise StreamlitDuplicateElementId(element_type)
```

The message names the fix: *"To fix this error, please pass a unique `key`
argument to the `<element_type>` element."* Two identical `st.button("ok")` calls
stop the script and show an error in the app — the opposite of a silent drop and
a misspelled console line.

Note what is *not* an error: two identical `st.write`s, two identical headings,
the same image twice. Those have no id, so they cannot collide.

### 7. The escape hatches

* **`key=`** on a widget makes its id user-controlled, which is how you get two
  otherwise identical widgets. On containers (`st.container(key=...)`) it also
  becomes the React key: `RenderNodeVisitor` uses `node.deltaBlock.id` when
  present "so that keyed containers maintain component identity across
  positional shifts (e.g. when a conditional element above the container causes
  it to move)".
* **`st.empty()`** returns a `DeltaGenerator` holding a `LockedCursor`, whose
  `lock_element()` returns itself. Writing to it repeatedly overwrites the same
  delta path. This is the supported way to mutate an already-emitted element,
  and it is the analogue of `toolgui`'s `NotifyPackUpdate` /
  `NotifyPackDelete`, which today only `ProgressBar` uses
  ([`progress_bar.go:35`](../../toolgui/tgcomp/tcmisc/progress_bar.go)).

### 8. React keys

`RenderNodeVisitor` (`frontend/lib/src/components/core/Block/RenderNodeVisitor.tsx`)
keys each child by `elementId || index.toString()` — the content-derived id when
the element has one, the ordinal otherwise. It also keeps a `Set` of keys used
in the current block and returns `null` for a repeat, with the comment that the
first is the live one because "our `setIn` logic pushes stale widgets down in
the list". That is a render-time backstop, not the identity mechanism; Python
has already raised by then.

## What this implies for `toolgui`

### The ordinal is available, but the protocol has to carry it

Issue #67 guesses "a per-container call ordinal is the usual answer", and
Streamlit confirms it. The Go side already has the natural place to count:
`Container.AddComponent` is the single funnel for every component
([`container.go:25`](../../toolgui/tgframe/container.go)). A counter on
`Container`, reset per run, reproduces `RunningCursor.index` almost exactly, and
`AddContainer` reproduces `open_block`.

The frontend is the harder half. `Forest` is a flat `{[id]: Node}` pool with
`parentID` back-pointers, and every operation is keyed by id. Two shapes are
possible:

1. **Keep the id as the key, derive it from the path.** The id becomes
   `container_component_main/3` (or `parentID + "_" + ordinal`), computed on the
   Go side and sent as it is today. `Forest` is untouched structurally;
   `createNode` starts hitting the "same id, update in place" branch it already
   has, and the insertion heuristic can be replaced with a real index. Cheapest,
   and it keeps `NotifyPack` as is.
2. **Send the path, address the tree by it.** Closer to Streamlit: `NotifyPack`
   gains `delta_path`, and `Forest` becomes a tree. This buys correct ordering
   by construction and kills the node pool (and its leak) outright, but it is a
   protocol change plus a rewrite of `Nodes.ts` and everything reading
   `forest.nodes`.

Both break every id currently visible in the DOM. `toolgui-e2e` selects on ids
like `#button_component_rerun`, and users can already pin ids with the
`WithID` / `Conf.ID` variants, so that escape hatch survives — but the defaults
change and that is a breaking change for anyone who wrote a selector or a
`GetClickID` comparison against a generated id.

### Widget state needs its own key, or the ordinal will move state around

If the id becomes positional, then state keyed by id becomes positional too, and
adding one component at the top of a page shifts every widget's state down by
one. Streamlit avoids this precisely by keeping the two identifiers separate:
placement is the ordinal, state is `$$ID-<hash>-<key>`.

So the ordinal change is not complete without deciding what `State` is keyed by.
The cheapest version of Streamlit's split for `toolgui`: keep the current
content-derived id as the *state* key (it is exactly a hash of the type and the
parameters), and give the node a separate positional key for the tree. The
duplicate check then applies to the state key only, and only for stateful
components.

### Duplicates should be loud

`tgframe.App.Run` already returns an error and the app shows it, so a duplicate
state key can surface the way `StreamlitDuplicateElementId` does, naming the
`WithID` variant as the fix. That is strictly better than the current console
line, whether or not the positional change happens.

### The lifecycle fix is independent of all of it

Issue #67 is right that the two halves of the `Forest` bug cancel out and must
land together. Streamlit's shape suggests the fix:

* Replace the `removing: boolean` with a run id stamped on each node, set from
  the create message and compared against the current run when pruning. No clear
  step to forget.
* Make the root containers structurally exempt, the way `ensureBlockNode` does,
  rather than by a membership test — the test is what is broken today, and a
  `Set` would fix this instance while leaving the same trap for the next root
  container someone adds.
* Do not prune when a run ends in error. `finishUpdate` currently prunes
  regardless of `pack.success`
  ([`App.tsx:143`](../../toolgui-web/lib/src/app/App.tsx)), so a page that
  panics halfway wipes everything it had not re-sent.

## Open questions

1. **Which of the two frontend shapes** — path-derived id string, or a real tree
   addressed by path. The first is a much smaller diff; the second is the one
   that removes the class of bug rather than this instance of it.
2. **What `State` is keyed by** once ids move. Nothing else in #67 can be
   designed until this is settled.
3. **Whether `Divider` and friends keep their zero-argument form.** With an
   ordinal they get a stable identity for free and `RandID` can be deleted. With
   any other scheme, `Divider` needs an argument, which the charts survey already
   decided for `Chart` ("the id is a required argument") — worth being
   consistent one way or the other.
4. **How much breakage is acceptable.** Every generated id in the DOM changes.
   Cheap mitigation: keep emitting the content-derived id as a stable
   `data-component-id` attribute for tests even after the tree key moves.
