import React, { Component } from "react";

import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import MenuIcon from '@mui/icons-material/Menu';

import { ThemeModeButton } from './ThemeModeButton';
import { AppConf } from "./AppConf";

interface AppNavbarProps {
  appConf: AppConf
  running: boolean
  pageFound: boolean
  pageName: string
  rerun: () => void
  onChange: (darkMode: string) => void
}

export class AppNavbar extends Component<AppNavbarProps> {
  constructor(props: AppNavbarProps) {
    super(props)
  }

  jumpToPage(name: string) {
    if (this.props.appConf.hash_page_name_mode) {
      window.location.href = '#/' + name
      window.location.reload();
    } else {
      window.location.href = '/' + name
    }
  }

  render() {
    return <AppBar position="static" color="default" sx={{ boxShadow: 0 }}>
      <Toolbar>
        {
          this.props.appConf.page_names.map(name =>
            <Button
              key={name}
              variant="text"
              color={name === this.props.pageName ? "primary" : "inherit"}
              onClick={() => { this.jumpToPage(name) }}
              startIcon={this.props.appConf.page_confs[name].emoji}>
              {this.props.appConf.page_confs[name].title}
            </Button>
          )
        }
      </Toolbar>
    </AppBar>
  }

  render1() {
    return <nav className="navbar" role="navigation" aria-label="main navigation">
      <div className="navbar-menu container">
        <div className="navbar-start">
          {
            this.props.appConf.page_names.map(name =>
              <a className={`navbar-item ${name === this.props.pageName ? 'is-active' : ''}`}
                onClick={() => { this.jumpToPage(name) }}>
                {this.props.appConf.page_confs[name].emoji}
                {this.props.appConf.page_confs[name].title}
              </a>
            )
          }
        </div>
        <div className="navbar-end">
          {this.props.running ?
            <div className="navbar-brand navbar-item">
              <span className="icon">
                <i className="fas fa-spinner fa-pulse"></i>
              </span>
            </div> : ''}
          <div className="buttons">
            {this.props.pageFound ?
              <button className="button navbar-item" onClick={() => { this.props.rerun() }}>
                Rerun
              </button> : ''}
            <ThemeModeButton onChange={(darkMode) => {
              this.props.onChange(darkMode)
            }} />
          </div>
        </div>
      </div>
    </nav>
  }
}