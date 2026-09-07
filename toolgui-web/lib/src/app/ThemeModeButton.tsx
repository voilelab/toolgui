import React from 'react'
import { ActionIcon, useComputedColorScheme, useMantineColorScheme } from '@mantine/core'
import { IconMoon, IconSun } from '@tabler/icons-react'

// ThemeModeButton flips Mantine's color scheme, which the rest of the app
// follows.
export function ThemeModeButton() {
  const { setColorScheme } = useMantineColorScheme()
  const dark = useComputedColorScheme('light', {
    getInitialValueInEffect: false,
  }) === 'dark'

  return (
    <ActionIcon variant="default" size="lg"
      aria-label={dark ? 'Switch to the light theme' : 'Switch to the dark theme'}
      onClick={() => { setColorScheme(dark ? 'light' : 'dark') }}>
      {dark ? <IconMoon size={18} /> : <IconSun size={18} />}
    </ActionIcon>
  )
}
