import { Node } from "../app/Nodes"
import { UpdateEvent } from "../app/UpdateEvent"
import { UploadFunc } from "../app/Upload"
import { DownloadFunc } from "../app/Download"
import { ThemeMode } from "../util/theme"

export interface Props {
  node: Node

  update: (event: UpdateEvent) => void
  upload: UploadFunc
  download: DownloadFunc

  // Page Theme
  theme: ThemeMode
}
