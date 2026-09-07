import React from "react"
import { Box, Text } from "@mantine/core"

import { Props } from "../component_interface"
import { emojize } from "../../util/emoji"

// Which way an arrow points, and the color a tone is painted in. The Go side
// decides both, so the rule for what counts as good news lives in one place.
const ARROWS: { [direction: string]: string } = { up: '↑', down: '↓' }
const TONES: { [tone: string]: string } = { positive: 'green', negative: 'red' }

export function TMetric({ node }: Props) {
  const direction: string = node.props.direction
  const tone: string = node.props.tone

  return (
    <Box id={node.props.id || undefined} mb="md">
      <Text size="sm" c="dimmed">{emojize(node.props.label)}</Text>
      <Text size="xl" fw={700}>{emojize(node.props.value)}</Text>
      {node.props.delta &&
        // The arrow carries the direction too, so the delta still reads for
        // someone who cannot tell the two colors apart.
        <Text size="sm" c={TONES[tone]}
          data-direction={direction} data-tone={tone}>
          {ARROWS[direction]} {node.props.delta}
        </Text>}
    </Box>
  )
}
