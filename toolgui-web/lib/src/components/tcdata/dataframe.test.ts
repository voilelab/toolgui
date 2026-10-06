import { expect, test } from 'vitest'

import { Cell, compare, display } from './dataframe'

const pct = (value: number | null, shown: string): Cell => ({ display: shown, value })

const sortBy = (cells: Cell[], type: "text" | "number" | "datetime", dir: number) =>
  cells.slice().sort((a, b) => compare(a, b, type, dir)).map(display)

test('a percentage sorts by its value, not its string', () => {
  const cells = [pct(0.2941, "29.41%"), pct(0.095, "9.5%"), pct(0.0021, "0.21%")]

  expect(sortBy(cells, "number", 1)).toEqual(["0.21%", "9.5%", "29.41%"])
  expect(sortBy(cells, "number", -1)).toEqual(["29.41%", "9.5%", "0.21%"])
})

test('a missing cell sorts last either way', () => {
  const cells = [pct(null, "-"), pct(2, "2"), pct(1, "1")]

  expect(sortBy(cells, "number", 1)).toEqual(["1", "2", "-"])
  expect(sortBy(cells, "number", -1)).toEqual(["2", "1", "-"])
})

test('bare strings still parse by column type', () => {
  expect(sortBy(["10", "9", "-"], "number", 1)).toEqual(["9", "10", "-"])
  expect(sortBy(["b", "a"], "text", 1)).toEqual(["a", "b"])
})

test('typed and bare cells mix in one column', () => {
  expect(sortBy(["3", pct(1, "1"), pct(null, "")], "number", 1))
    .toEqual(["1", "3", ""])
})
