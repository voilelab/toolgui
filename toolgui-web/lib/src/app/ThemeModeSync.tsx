import React, { ReactNode, useEffect } from 'react'
import { useComputedColorScheme } from '@mantine/core'

import { ThemeMode, applyBulmaTheme } from '../util/theme'

interface ThemeModeSyncProps {
  children: (themeMode: ThemeMode) => ReactNode
}

// ThemeModeSync hands Mantine's color scheme to the parts of the app that
// draw themselves rather than read CSS, and mirrors it onto <html> for Bulma.
export function ThemeModeSync({ children }: ThemeModeSyncProps) {
  // getInitialValueInEffect off: the scheme is known on the first render, so
  // there is nothing to wait for and no flash of the wrong theme.
  const themeMode = useComputedColorScheme('light', {
    getInitialValueInEffect: false,
  })

  useEffect(() => { applyBulmaTheme(themeMode) }, [themeMode])

  return <>{children(themeMode)}</>
}
