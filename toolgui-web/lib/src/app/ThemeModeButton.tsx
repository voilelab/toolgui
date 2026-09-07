import React from 'react'
import { useComputedColorScheme, useMantineColorScheme } from '@mantine/core'

// ThemeModeButton flips Mantine's color scheme, which the rest of the app
// follows.
export function ThemeModeButton() {
  const { setColorScheme } = useMantineColorScheme()
  const dark = useComputedColorScheme('light', {
    getInitialValueInEffect: false,
  }) === 'dark'

  return (
    <button className="button"
      onClick={() => { setColorScheme(dark ? 'light' : 'dark') }}>
      <span className="icon">
        {dark ?
          <i className="fas fa-moon"></i> :
          <i className="fas fa-sun"></i>}
      </span>
    </button>
  )
}
