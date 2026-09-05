import { render, screen } from '@testing-library/react'
import { expect, test } from 'vitest'

// Smoke test: proves the Vitest + jsdom + React toolchain is wired up.
test('renders the app root', () => {
  render(<div id="root">toolgui</div>)
  expect(screen.getByText('toolgui')).toBeInTheDocument()
})
