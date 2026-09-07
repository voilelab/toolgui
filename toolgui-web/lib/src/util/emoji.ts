import shortcodesData from 'emojibase-data/en/shortcodes/github.json'

// emojibase-data's type sidecar pulls its type from the emojibase peer, which
// toolgui does not install, so the import lands as any. Name the shape here.
const shortcodes: Record<string, string | string[]> = shortcodesData

// emojibase keys by hexcode ("1F600", "1F1E6-1F1E8") and holds either one
// shortcode or several, so the lookup a renderer wants has to be inverted out
// of it. The github preset is the one people have muscle memory for, from
// GitHub and Slack; see docs/design/emoji-shortcode.md for why not the wider
// sets.
const table: Record<string, string> = Object.create(null)
for (const [hexcode, value] of Object.entries(shortcodes)) {
  const emoji = String.fromCodePoint(
    ...hexcode.split('-').map((point) => parseInt(point, 16)))
  for (const shortcode of Array.isArray(value) ? value : [value]) {
    if (!(shortcode in table)) {
      table[shortcode] = emoji
    }
  }
}

// A shortcode is a name between colons, as on GitHub and Slack. Digits alone
// are a name too, :100: and :1234:, so the table lookup rather than the shape
// of the name is what decides: the seconds in a time like 10:30:00 survive
// because "30" is not in the table.
const SHORTCODE = /:([a-z0-9_+-]+):/g

// emojize replaces the emoji shortcodes in text with the emoji they stand
// for. A name the table does not know stays as it was written.
export function emojize(text: string): string {
  if (text.indexOf(':') === -1) {
    return text
  }

  return text.replace(SHORTCODE, (shortcode, name) =>
    name in table ? table[name] : shortcode)
}
