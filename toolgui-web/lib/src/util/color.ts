// Go sends its own colour names (tcutil.Color) over the wire. Mantine names
// its palette differently, so translate at the edge rather than teaching the
// Go side about either.
const MANTINE_COLORS: { [key: string]: string } = {
  info: 'blue',
  success: 'green',
  warning: 'yellow',
  danger: 'red',
}

// mantineColor is the Mantine palette key for a Go colour. Undefined for the
// empty colour, which leaves a component on its neutral default.
export function mantineColor(color: string | undefined): string | undefined {
  return color ? MANTINE_COLORS[color] : undefined
}

// inputBorderStyles paints an input's border in a Go colour, which is what
// the colour meant before: a mark on the field, not on its label.
export function inputBorderStyles(color: string | undefined) {
  const name = mantineColor(color)
  if (!name) {
    return undefined
  }

  return { input: { borderColor: `var(--mantine-color-${name}-filled)` } }
}
