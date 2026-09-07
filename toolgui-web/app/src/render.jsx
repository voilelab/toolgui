import React from 'react'
import { render as rtlRender } from '@testing-library/react'
import { MantineProvider } from '@mantine/core'

// The components are Mantine's now, so a test has to give them the provider
// the app gives them.
export function render(ui, options) {
  return rtlRender(ui, { wrapper: MantineProvider, ...options })
}
