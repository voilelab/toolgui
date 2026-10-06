import React, { useEffect, useRef } from "react"
import { Input } from "@mantine/core"

import { Compartment, EditorState } from "@codemirror/state"
import {
  EditorView, drawSelection, highlightActiveLine, highlightActiveLineGutter,
  keymap, lineNumbers,
} from "@codemirror/view"
import {
  defaultKeymap, history, historyKeymap, indentWithTab,
} from "@codemirror/commands"
import {
  bracketMatching, defaultHighlightStyle, indentOnInput, syntaxHighlighting,
} from "@codemirror/language"
import { closeBrackets, closeBracketsKeymap } from "@codemirror/autocomplete"
import { highlightSelectionMatches, searchKeymap } from "@codemirror/search"
import { oneDarkHighlightStyle } from "@codemirror/theme-one-dark"

import { stateValues, syncResetKey } from "../state"
import { Props } from "../component_interface"
import { ThemeMode } from "../../util/theme"
import { loadLanguage } from "./code_lang"
import { codeInputBoxStyle } from "./code_input"

// Colors come from Mantine, so the editor follows the page's light/dark mode.
const editorTheme = EditorView.theme({
  "&": {
    height: "100%",
    color: "var(--mantine-color-text)",
    backgroundColor: "var(--mantine-color-body)",
    border: "1px solid var(--mantine-color-default-border)",
    borderRadius: "var(--mantine-radius-default)",
  },
  "&.cm-focused": {
    outline: "none",
    borderColor: "var(--mantine-primary-color-filled)",
  },
  ".cm-scroller": {
    fontFamily: "var(--mantine-font-family-monospace)",
    lineHeight: "1.5",
  },
  ".cm-gutters": {
    color: "var(--mantine-color-dimmed)",
    backgroundColor: "var(--mantine-color-default)",
    borderRight: "1px solid var(--mantine-color-default-border)",
  },
  ".cm-activeLine, .cm-activeLineGutter": {
    backgroundColor: "var(--mantine-color-default-hover)",
  },
})

function themeExtension(theme: ThemeMode) {
  return [
    EditorView.darkTheme.of(theme === "dark"),
    syntaxHighlighting(
      theme === "dark" ? oneDarkHighlightStyle : defaultHighlightStyle),
  ]
}

// CodeEditor is loaded apart from the main bundle, see code_input.tsx.
export default function CodeEditor({ node, update, theme }: Props) {
  const id: string = node.props.id
  const labelId = `${id}-label`

  const host = useRef<HTMLDivElement>(null)
  const view = useRef<EditorView | null>(null)
  const langConf = useRef(new Compartment())
  const themeConf = useRef(new Compartment())

  // The latest update, for the handlers the editor was built with.
  const updateRef = useRef(update)
  updateRef.current = update

  useEffect(() => {
    syncResetKey(id, node.props.reset_key)
    const initial: string = stateValues[id] ?? node.props.default ?? ""

    // Send only what changed since the last send, so a blur alone does not
    // rerun the page.
    let dirty = false
    const send = (v: EditorView) => {
      if (!dirty) {
        return
      }
      dirty = false
      updateRef.current({
        type: "input",
        id: id,
        value: v.state.doc.toString(),
      })
    }

    const v = new EditorView({
      parent: host.current!,
      state: EditorState.create({
        doc: initial,
        extensions: [
          lineNumbers(),
          highlightActiveLineGutter(),
          history(),
          drawSelection(),
          indentOnInput(),
          bracketMatching(),
          closeBrackets(),
          highlightActiveLine(),
          highlightSelectionMatches(),
          keymap.of([
            { key: "Mod-Enter", run: (v) => { send(v); return true } },
            ...closeBracketsKeymap,
            ...defaultKeymap,
            ...searchKeymap,
            ...historyKeymap,
            // Esc then Tab moves focus out, so Tab is no keyboard trap.
            indentWithTab,
          ]),
          editorTheme,
          themeConf.current.of(themeExtension(theme)),
          langConf.current.of([]),
          EditorView.contentAttributes.of({
            id: id,
            "aria-labelledby": labelId,
          }),
          EditorView.updateListener.of((u) => {
            if (u.docChanged) {
              stateValues[id] = u.state.doc.toString()
              dirty = true
            }
          }),
          EditorView.domEventHandlers({
            blur: (_, v) => { send(v) },
          }),
        ],
      }),
    })
    view.current = v

    return () => {
      v.destroy()
      view.current = null
    }
  }, [])

  useEffect(() => {
    view.current?.dispatch({
      effects: themeConf.current.reconfigure(themeExtension(theme)),
    })
  }, [theme])

  useEffect(() => {
    let stale = false
    const reconfigure = (ext: any) => {
      if (!stale) {
        view.current?.dispatch({ effects: langConf.current.reconfigure(ext) })
      }
    }

    const loading = loadLanguage(node.props.lang)
    if (loading) {
      loading.then(reconfigure, () => reconfigure([]))
    } else {
      reconfigure([])
    }

    return () => { stale = true }
  }, [node.props.lang])

  return (
    <Input.Wrapper
      label={node.props.label}
      labelElement="div"
      labelProps={{
        id: labelId,
        onClick: () => view.current?.focus(),
      }}
      mb="md">
      <div ref={host} className="toolgui-code-input"
        style={codeInputBoxStyle(node.props.height)} />
    </Input.Wrapper>
  )
}
