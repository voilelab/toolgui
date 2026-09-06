import { Node } from "../app/Nodes"
import { UpdateEvent } from "../app/UpdateEvent"
import { UploadFunc } from "../app/Upload"
import { ThemeMode } from "../util/theme"

export interface Props {
  node: Node

  update: (event: UpdateEvent) => void
  upload: UploadFunc

  // Page Theme
  theme: ThemeMode
}
