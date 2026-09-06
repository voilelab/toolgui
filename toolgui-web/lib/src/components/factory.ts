import { TTextbox } from "./tcinput/textbox"
import { TCheckbox } from "./tcinput/checkbox"
import { TButton } from "./tcinput/button"
import { TSelect } from "./tcinput/select"
import { TTextarea } from "./tcinput/textarea"
import { TFileupload } from "./tcinput/fileupload"
import { TRadio } from "./tcinput/radio"
import { TDatepicker } from "./tcinput/datepicker"
import { TNumber } from "./tcinput/number"
import { TForm } from "./tcinput/form"
import { TDownloadButton } from "./tcinput/download_button"

import { TContainer } from "./tclayout/container"
import { TBox } from "./tclayout/box"
import { TColumn } from "./tclayout/column"

import { TTitle } from "./tccontent/title"
import { TImage } from "./tccontent/image"
import { TSubtitle } from "./tccontent/subtitle"
import { TText } from "./tccontent/text"
import { TDivider } from "./tccontent/divider"
import { TMarkdown } from "./tccontent/markdown"
import { TCode } from "./tccontent/code"
import { TLink } from "./tccontent/link"
import { TIframe } from "./tcmisc/iframe"

import { TJson } from "./tcdata/json"
import { TTable } from "./tcdata/table"
import { TChart } from "./tcdata/chart"

import { TProgressar } from "./tcmisc/progress_bar"
import { TMessage } from "./tcmisc/message"
import { THtml } from "./tcmisc/html"

import { Props } from "./component_interface"
import { TTab } from "./tclayout/tab"
import { TLatex } from "./tccontent/latex"
import { TExpand } from "./tclayout/expand"

const creatorMap: { [id: string]: ((props: Props) => JSX.Element) } = {
  textbox_component: TTextbox,
  checkbox_component: TCheckbox,
  button_component: TButton,
  select_component: TSelect,
  textarea_component: TTextarea,
  fileupload_component: TFileupload,
  radio_component: TRadio,
  datepicker_component: TDatepicker,
  number_component: TNumber,
  form_component: TForm,
  download_button_component: TDownloadButton,

  container_component: TContainer,
  box_component: TBox,
  column_component: TColumn,
  tab_component: TTab,
  expand_component: TExpand,

  title_component: TTitle,
  subtitle_component: TSubtitle,
  image_component: TImage,
  text_component: TText,
  divider_component: TDivider,
  markdown_component: TMarkdown,
  code_component: TCode,
  link_component: TLink,
  latex_component: TLatex,

  json_component: TJson,
  table_component: TTable,
  chart_component: TChart,

  progress_bar_component: TProgressar,
  message_component: TMessage,
  iframe_component: TIframe,
  html_component: THtml,
}


export function TComponent(props: Props) {
  const name = props.node.props.name
  if (!(name in creatorMap)) {
    throw new Error(`unsupported component type: ${name}`);
  }

  return creatorMap[name](props)
}