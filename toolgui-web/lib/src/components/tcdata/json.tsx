import React, { useState } from "react"

import { Props } from "../component_interface"

import '@toolgui-web/lib/src/assets/css/json.css'


export function TJson({ node, theme }: Props) {
  return (
    <div id={node.props.id || undefined} className="block">
      <pre>
        <JsonValue val={JSON.parse(node.props.value)} pad={0} theme={theme} />
      </pre>
    </div>
  )
}

function JsonValue({ val, pad, theme }: { val: any, pad: number, theme: string }) {
  if (val === null) {
    return JsonNull({ val, theme })
  }

  if (Array.isArray(val)) {
    return JsonList({ val, pad, theme })
  } else if (typeof val === 'object') {
    return JsonDict({ val, pad, theme })
  } else if (typeof val === 'string') {
    return JsonString({ val, theme })
  } else if (typeof val === 'boolean') {
    return JsonBool({ val, theme })
  } else {
    return JsonElement({ val, theme })
  }
}

function JsonDict({ val, pad, theme }: { val: any, pad: number, theme: string }) {
  const [open, setOpen] = useState(true)
  if (!open) {
    return <p style={{ display: 'inline' }}>
      <span className="pseudolink" onClick={(e) => { setOpen(true) }}>
        <b>{"{ ... }"}</b>
      </span>
    </p>
  }

  const kvs = []
  for (const [key, value] of Object.entries(val)) {
    kvs.push(<div>{' '.repeat(pad + 2)}"{key}": <JsonValue val={value} pad={pad + 2} theme={theme} />,</div>)
  }
  return (
    <p style={{ display: 'inline' }}>
      <span className="pseudolink" onClick={(e) => { setOpen(false) }}>
        <b>{"{"}</b>
      </span>
      <div>{kvs}</div>
      <span className="pseudolink" onClick={(e) => { setOpen(false) }}>
        <b>{' '.repeat(pad)}{"}"}</b>
      </span>
    </p>
  )
}

function JsonList({ val, pad, theme }: { val: { [key: string]: any }, pad: number, theme: string }) {
  const [open, setOpen] = useState(true)
  if (!open) {
    return <p style={{ display: 'inline' }}>
      <span className="pseudolink" onClick={(e) => { setOpen(true) }}>
        <b>{"[ ... ]"}</b>
      </span>
    </p>
  }

  return (
    <p style={{ display: 'inline' }}>
      <span className="pseudolink" onClick={(e) => { setOpen(false) }}>
        <b>{"["}</b>
      </span>
      <div>{
        val.map((value: any) => (
          <div>{' '.repeat(pad + 2)}<JsonValue val={value} pad={pad + 2} theme={theme} />,</div>
        ))
      }</div>
      <span className="pseudolink" onClick={(e) => { setOpen(false) }}>
        <b>{' '.repeat(pad)}{"]"}</b>
      </span>
    </p>
  )
}

function JsonString({ val, theme }: { val: string, theme: string }) {
  return <p style={{ display: 'inline', color: theme === 'light' ? 'green' : 'lightgreen' }} >"{val}"</p>
}

function JsonBool({ val, theme }: { val: boolean, theme: string }) {
  const strVal = val ? 'true' : 'false'
  return <p style={{ display: 'inline', color: theme === 'light' ? 'brown' : 'lightcoral' }} >{strVal}</p>
}

function JsonNull({ val, theme }: { val: null, theme: string }) {
  return <p style={{ display: 'inline', color: theme === 'light' ? 'brown' : 'lightcoral' }} >null</p>
}

function JsonElement({ val, theme }: { val: any, theme: string }) {
  return <p style={{ display: 'inline', color: theme === 'light' ? 'brown' : 'lightcoral' }}>{val}</p>
}
