import React, { ReactNode } from 'react'
import { useComputedColorScheme } from '@mantine/core'

import { ThemeMode } from '../util/theme'

interface ThemeModeSyncProps {
  children: (themeMode: ThemeMode) => ReactNode
}

// ThemeModeSync hands Mantine's color scheme, the app's one source of truth
// for the theme, to the parts of the app that draw themselves rather than
// read CSS.
export function ThemeModeSync({ children }: ThemeModeSyncProps) {
  // getInitialValueInEffect off: the scheme is known on the first render, so
  // there is nothing to wait for and no flash of the wrong theme.
  const themeMode = useComputedColorScheme('light', {
    getInitialValueInEffect: false,
  })

  return <>{children(themeMode)}</>
}
