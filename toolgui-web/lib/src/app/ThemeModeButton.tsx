import React, { Component } from 'react'

import { ThemeMode } from '../util/theme'

interface ThemeModeButtonProps {
  darkMode: ThemeMode
  onChange: (darkMode: ThemeMode) => void
}

// ThemeModeButton only reports the flip; the app owns the theme, so there is
// one place that decides what it started as.
export class ThemeModeButton extends Component<ThemeModeButtonProps> {
  render() {
    const dark = this.props.darkMode === 'dark'

    return (
      <button className="button"
        onClick={() => { this.props.onChange(dark ? 'light' : 'dark') }}>
        <span className="icon">
          {dark ?
            <i className="fas fa-moon"></i> :
            <i className="fas fa-sun"></i>}
        </span>
      </button>
    )
  }
}
