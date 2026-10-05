import React, { useState } from "react"
import { Anchor, Divider, Modal, Stack, Text, UnstyledButton, getDefaultZIndex } from "@mantine/core"

import { AppConf } from "./AppConf"
import { MarkdownText } from "../components/tccontent/markdown"
import { useEscapeToClose, useOverlay } from "../components/tclayout/overlay_stack"
import { ThemeMode } from "../util/theme"

const REPO_URL = "https://github.com/voilelab/toolgui"
const DOCS_URL = "https://voilelab.github.io/toolgui/"

// Not a node id, so it cannot clash with a page's dialogs on the stack.
const OVERLAY_ID = "toolgui-about"

interface AboutProps {
  appConf: AppConf
  theme: ThemeMode
}

// AboutLine is the side nav's version line, a button opening the About
// dialog. Without the version it reads "About", and is left out when there
// is nothing to show.
export function AboutLine({ appConf, theme }: AboutProps) {
  const [opened, setOpened] = useState(false)

  const showVersion = appConf.show_version
  const about = appConf.about || ""
  if (!showVersion && !about) {
    return null
  }

  return <>
    <UnstyledButton className="toolgui-nav-version" onClick={() => { setOpened(true) }}>
      <Text size="xs" c="dimmed" span>
        {showVersion ? `toolgui ${appConf.version}` : "About"}
      </Text>
    </UnstyledButton>
    <AboutDialog opened={opened} close={() => { setOpened(false) }}
      about={about} version={showVersion ? appConf.version : ""} theme={theme} />
  </>
}

function AboutDialog({ opened, close, about, version, theme }: {
  opened: boolean
  close: () => void
  about: string
  version: string
  theme: ThemeMode
}) {
  // On the overlay stack like a page's Dialog, so ESC closes only the top one.
  const { depth, isTop } = useOverlay(OVERLAY_ID, "dialog", opened)
  useEscapeToClose(opened && isTop, close)

  const zIndex = Math.min(
    getDefaultZIndex("modal") + Math.max(depth, 0),
    getDefaultZIndex("popover") - 1)

  return (
    <Modal className="toolgui-about" title="About" opened={opened} onClose={close}
      zIndex={zIndex} closeOnEscape={false}
      closeButtonProps={{ "aria-label": "Close dialog" }}>
      <Stack gap="sm">
        {about ? <MarkdownText text={about} theme={theme} /> : null}
        {about && version ? <Divider /> : null}
        {version ?
          <Stack gap={4} className="toolgui-about-toolgui">
            <Text size="sm" fw={500}>toolgui {version}</Text>
            <Text size="sm" c="dimmed">
              A Go framework for building interactive dashboards and web apps,
              in the spirit of Streamlit.
            </Text>
            <Anchor size="sm" href={REPO_URL} target="_blank">{REPO_URL}</Anchor>
            <Anchor size="sm" href={DOCS_URL} target="_blank">Documentation</Anchor>
          </Stack> : null}
      </Stack>
    </Modal>
  )
}
