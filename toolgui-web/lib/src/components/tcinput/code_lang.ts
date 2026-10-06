import type { Extension } from "@codemirror/state"
import { StreamLanguage } from "@codemirror/language"

// Each language is its own chunk, loaded on first use.
const loaders: { [lang: string]: () => Promise<Extension> } = {
  go: () => import("@codemirror/lang-go").then((m) => m.go()),
  python: () => import("@codemirror/lang-python").then((m) => m.python()),
  javascript: () =>
    import("@codemirror/lang-javascript").then((m) => m.javascript()),
  typescript: () => import("@codemirror/lang-javascript")
    .then((m) => m.javascript({ typescript: true })),
  jsx: () => import("@codemirror/lang-javascript")
    .then((m) => m.javascript({ jsx: true })),
  tsx: () => import("@codemirror/lang-javascript")
    .then((m) => m.javascript({ jsx: true, typescript: true })),
  json: () => import("@codemirror/lang-json").then((m) => m.json()),
  sql: () => import("@codemirror/lang-sql").then((m) => m.sql()),
  html: () => import("@codemirror/lang-html").then((m) => m.html()),
  css: () => import("@codemirror/lang-css").then((m) => m.css()),
  markdown: () => import("@codemirror/lang-markdown").then((m) => m.markdown()),
  yaml: () => import("@codemirror/lang-yaml").then((m) => m.yaml()),
  shell: () => import("@codemirror/legacy-modes/mode/shell")
    .then((m) => StreamLanguage.define(m.shell)),
}

const aliases: { [lang: string]: string } = {
  golang: "go", py: "python", js: "javascript", ts: "typescript",
  md: "markdown", yml: "yaml", sh: "shell", bash: "shell",
}

// loadLanguage resolves to the highlighting for lang, or null when lang is
// not supported, which leaves the code as plain text.
export function loadLanguage(lang: string | undefined): Promise<Extension> | null {
  const key = (lang ?? "").toLowerCase()
  const loader = loaders[aliases[key] ?? key]
  return loader ? loader() : null
}
