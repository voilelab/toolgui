import { Node } from "../app/Nodes"
import { UpdateEvent } from "../app/UpdateEvent"
import { UploadFunc } from "../app/Upload"

export interface Props {
  node: Node

  update: (event: UpdateEvent) => void
  upload: UploadFunc

  // Page Theme (light or dark)
  theme: string
}
