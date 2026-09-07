package tgcomp

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tccontent"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// Base is embedded in every component conf, and is where a conf's ID comes
// from. A third-party component's conf embeds it the same way the built-in
// ones do.
type Base = tgframe.Base

// Badge show a short label, for a status or a tag next to other content.
var Badge = tccontent.Badge

// BadgeConf is the configuration for the Badge component.
type BadgeConf = tccontent.BadgeConf

// Caption show a small dimmed text, for a note next to what it explains.
var Caption = tccontent.Caption

// CaptionConf is the configuration for the Caption component.
type CaptionConf = tccontent.CaptionConf

// Code create a code block with syntax highlight.
var Code = tccontent.Code

// CodeConf is the configuration for the Code component.
type CodeConf = tccontent.CodeConf

// Divider create a horizontal line.
var Divider = tccontent.Divider

// DividerConf is the configuration for the Divider component.
type DividerConf = tccontent.DividerConf

// Image show a image.
var Image = tccontent.Image

// ImageConf is the configuration for the Image component.
type ImageConf = tccontent.ImageConf

// Link create a link component.
var Link = tccontent.Link

// LinkConf is the configuration for the Link component.
type LinkConf = tccontent.LinkConf

// LinkButton create a link that is drawn as a button.
var LinkButton = tccontent.LinkButton

// LinkButtonConf is the configuration for the LinkButton component.
type LinkButtonConf = tccontent.LinkButtonConf

// Markdown render markdown to html.
var Markdown = tccontent.Markdown

// MarkdownConf is the configuration for the Markdown component.
type MarkdownConf = tccontent.MarkdownConf

// Metric show a labelled value, with an optional delta under it.
var Metric = tccontent.Metric

// MetricConf is the configuration for the Metric component.
type MetricConf = tccontent.MetricConf

// Subtitle create a subtitle.
var Subtitle = tccontent.Subtitle

// SubtitleConf is the configuration for the Subtitle component.
type SubtitleConf = tccontent.SubtitleConf

// Text show a text.
var Text = tccontent.Text

// TextConf is the configuration for the Text component.
type TextConf = tccontent.TextConf

// Title show a title.
var Title = tccontent.Title

// TitleConf is the configuration for the Title component.
type TitleConf = tccontent.TitleConf

// Latex create a latex component.
var Latex = tccontent.Latex

// LatexConf is the configuration for the Latex component.
type LatexConf = tccontent.LatexConf
