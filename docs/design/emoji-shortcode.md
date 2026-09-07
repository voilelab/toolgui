# Emoji shortcodes in content

Survey for [#17](https://github.com/voilelab/toolgui/issues/17): let an app
author write `:tada:` and have the user see 🎉.

Nothing here is implemented yet. This is the shape of the decision and a
recommendation.

## Where text flows today

A content component carries its text as a plain JSON prop from Go to React,
untouched on both ends:

* `toolgui/tgcomp/tccontent/text.go:12` — `Text string` with a `json:"text"` tag,
  same for `title.go`, `subtitle.go`, `link.go`, `markdown.go`.
* `toolgui-web/lib/src/components/tccontent/text.tsx` renders
  `{node.props.text}` into a `div`; `title.tsx` and `subtitle.tsx` do the same
  into `h1`/`h2`.
* `markdown.tsx` is the exception: it feeds the string to `react-markdown`,
  so it already has a remark pipeline we can extend.

Beyond `tccontent`, the same shape shows up in input labels
(`tcinput/*.go`, `json:"label"`), `tclayout/expand.go` (`json:"title"`),
`tcmisc/message.go`, and `tgframe/app.go:50` — `PageConfig.Emoji`, which the
side nav (`AppSideNav.tsx:98`) and the favicon (`util/seticon.ts`) render as a
literal emoji character today.

So there is exactly one place a shortcode could be expanded on the Go side
(component construction) and exactly one on the web side (the shared
`toolgui-web/lib`, which both the browser app and the Wails app consume).

## Decision 1: which layer expands the shortcode

### A. Go, at component construction

`Text()` runs the string through a replacer before storing it.

* One table, and app authors could reuse the helper on their own strings.
* `PageConfig.Emoji` gets shortcodes for free.
* No browser bundle cost.
* But `go.mod` has exactly one dependency today (`golang.org/x/net`). An
  emoji library (`kyokomi/emoji`, `enescakir/emoji`) breaks that, and a
  generated table is ~42 KB of Go source (github preset) linked into every
  binary, `tccontent` importers or not.
* The blocker: **Go cannot see markdown structure.** Expanding `:smile:`
  inside a fenced code block in `tgcomp.Markdown` is wrong, and a
  string-level replacer cannot tell the difference. Same for
  `tgcomp.Code` — a code component is the one place a shortcode must stay
  literal.

### B. Web, at render (recommended)

`toolgui-web/lib` expands shortcodes as it renders.

* The markdown path can be markdown-aware, so code spans and fences stay
  literal for free.
* `go.mod` stays at one dependency; Go binaries do not grow.
* Side nav and favicon are already rendered here, so `PageConfig.Emoji`
  still gets shortcode support without touching Go.
* Cost: bundle size (see Decision 2) and a per-component opt-in — every
  component that should expand needs the call added.

### C. Neither — an explicit opt-in helper

Ship `tgutil.Emojize(string) string` and nothing automatic. Authors write
`tgcomp.Text(p.Main, tgutil.Emojize("hi :smile:"))`.

* Zero surprise, zero behaviour change, smallest diff.
* But it is not really the feature: the point of a shortcode is that you
  type it inline. It also still pays the Go table cost of option A.

Worth keeping as the fallback if we decide automatic expansion is too much
magic.

## Decision 2: which dataset

[emojibase](https://emojibase.dev/docs/datasets/) (the link in the issue)
ships several shortcode presets under
`emojibase-data/en/shortcodes/`. Measured on `emojibase-data@17.0.0`, after
inverting to a `shortcode -> emoji` map and dropping everything else:

| preset | shortcodes | raw | gzipped |
| --- | --- | --- | --- |
| `github.json` | 1913 | 42 KB | 15 KB |
| `emojibase.json` | 5055 | 161 KB | 47 KB |
| `joypixels.json` | — | larger still | — |

The data is keyed by hexcode (`"1F600"`, `"1F1E6-1F1E8"`) with a
`string | string[]` value, so building the map we want is a few lines:

```ts
const emoji = hexcode.split('-').map(h => parseInt(h, 16))
String.fromCodePoint(...emoji)
```

`github` is the recommendation. It is what people have muscle memory for
from GitHub and Slack, and it is a quarter of the size. `emojibase`'s wider
set uses names that differ from the ones authors will actually type.

No preset is the map we want, though. emojibase keys shortcodes by hexcode
and keeps the emoji character in a separate file (`en/compact.json`), so
`shortcode -> emoji` has to be inverted out of the dataset either way.
emojibase's own answer to that is `fetchShortcodes` against its CDN, which
is not open to us: a toolgui app has to work on a machine with no route to
the internet, as `App.tsx` says in its first comment.

So the question is who does the inverting. Three ways, measured by bundling
each with esbuild, minified and gzipped:

| approach | raw | gzipped |
| --- | --- | --- |
| `node-emoji`'s `emojify()` | 227 KB | 43 KB |
| a generated `shortcode -> emoji` table checked in | 51 KB | 16 KB |
| import the preset, invert it at module load | 46 KB | 16 KB |

`node-emoji` is the least code — `emojify(':tada:')` is the whole feature,
and it leaves `15:04:05` alone — but 2.7x the gzip, because `emojilib`
carries the keyword data behind `find`, `search`, and `which`, none of
which we call. `remark-gemoji` (12 KB) has the same problem one level down:
it pulls `gemoji`, a single non-tree-shakeable 347 KB / 64 KB module of
descriptions, tags, and categories.

The last two cost the same, so the tiebreaker is what lands in the
repository. Inverting at module load is ~10 lines against a pinned
dependency; the generated table is 50 KB of checked-in source plus a script
to regenerate it and a way to notice when it has drifted. Take the ten
lines.

One wrinkle either way: `lib` is built by Babel with no bundler
(`toolgui-web/lib/package.json`, `babel --out-dir build`), so the JSON
import is left alone and resolved by the *consumer's* Vite. That works for
the browser app, the Wails app, and vitest. TypeScript resolves the import
through emojibase-data's type sidecar, which pulls its type from the
`emojibase` peer we do not install, so the import lands as `any` — name the
shape at the import site rather than adding the peer.

## Decision 3: what expands

Proposed scope, tightest first:

1. **Expands**: `Title`, `Subtitle`, `Text`, `Link` text, `Markdown`
   (text nodes only), `PageConfig.Emoji`.
2. **Also reasonable, later**: input `Label`s, `Expand` title, `Message`
   title, `Tab` labels. Same helper, more call sites.
3. **Never expands**: `Code` (`tccontent/code.go`), code spans and fences
   inside markdown, `Html`, `Json`, `Table` cells. A shortcode there is
   data, not decoration.

Values that round-trip through state — what a user typed into a `Textbox`
and the app echoes back — are worth a decision. Expanding them makes
`Text(p.Main, state.String("box"))` behave differently from the raw state
value the app sees, which is confusing. Recommend: expansion is a rendering
concern only, so the echo does expand and the Go-side value does not. Say
so in the docs.

## Edge cases

* **Unknown shortcodes stay literal.** `:notanemoji:` renders as typed. Do
  not strip.
* **False positives.** A regex of `/:([a-z0-9_+-]+):/g` over prose will hit
  things like `10:30:00` — `30` is not a shortcode so it survives, but the
  pattern is worth a test. Requiring at least one non-digit in the name
  looked like a cheap extra guard until the data said otherwise: the github
  preset has `100` and `1234`, and `:100:` is one of the shortcodes people
  reach for most. The table lookup is the only filter, then, and a false
  positive needs prose where a numeric run happens to *be* a shortcode
  (`a:100:b`). Rare enough to accept.
* **Escaping.** No escape hatch in v1; `Code`/markdown code is the escape
  hatch. If one is wanted later, `\:smile\:` is the conventional spelling.
* **Skin tone modifiers.** `:+1::skin-tone-2:` is a Slack-ism the github
  preset does not cover. Out of scope for v1.
* **Presentation selectors.** Some hexcodes carry `FE0F`; expanding from
  the hexcode key preserves whatever the dataset has, which is what we
  want.

## Recommendation

1. Take `emojibase-data` as a dependency of `toolgui-web/lib`.
2. Add `toolgui-web/lib/src/util/emoji.ts`: invert
   `emojibase-data/en/shortcodes/github.json` into a `shortcode -> emoji`
   table at module load, and export
   `emojize(text: string): string`, a single regex pass over it.
3. Call it from `title.tsx`, `subtitle.tsx`, `text.tsx`, `link.tsx`,
   `AppSideNav.tsx`, and the `setIcon` path.
4. For `markdown.tsx`, add a small local remark plugin that walks `text`
   nodes only (so `code`/`inlineCode` are skipped) and applies the same
   table — same data as the plain path, no `gemoji` dependency.
5. Document it in `docs/src/components/content/`, and add a Cypress case in
   `toolgui-e2e` asserting `:tada:` renders as 🎉 in `Text` and stays literal
   in `Code`.

**Status: all five shipped.** The table carries 1913 shortcodes and costs the
app bundle 45 KB raw / 16 KB gzipped, measured by building it with and
without the change. Step 4 landed as `util/remark_emoji.ts`, a visit of
`text` nodes, which is what keeps code spans and fences literal. The
user-facing documentation is `docs/src/components/content/emoji.md`.

Two things the work changed here. The digit guard in the edge cases above is
gone: the github preset has all-digit names, so it would have dropped
`:100:`. And this survey called for checking a generated table into the
repository, which measured the same as inverting at load and so bought
nothing.

If instead we want the Go side to own this, option C (`tgutil.Emojize`) is
the honest version of it — automatic Go-side expansion cannot get markdown
right.
