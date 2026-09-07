// Go sends Bulma's colour names (tcutil.Color) over the wire. Mantine names
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
