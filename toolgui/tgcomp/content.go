package tgcomp

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tccontent"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// Base is embedded in every component conf, and is where a conf's ID comes
// from. A third-party component's conf embeds it the same way the built-in
// ones do.
type Base = tgframe.Base

// Code create a code block with syntax highlight.
var Code = tccontent.Code

type CodeConf = tccontent.CodeConf

// CodeWithConf create a code block with syntax highlight.
var CodeWithConf = tccontent.CodeWithConf

// Divider create a horizontal line.
var Divider = tccontent.Divider

// DividerWithID create a horizontal line with ID.
var DividerWithID = tccontent.DividerWithID

// Image show a image.
var Image = tccontent.Image

// ImageConf is the configuration for the Image component
type ImageConf = tccontent.ImageConf

// ImageWithConf show a image with a custom configuration.
var ImageWithConf = tccontent.ImageWithConf

// Link create a link component.

var Link = tccontent.Link

// LinkWithID create a link component with a user specific id.
var LinkWithID = tccontent.LinkWithID

// Markdown render markdown to html.
var Markdown = tccontent.Markdown

// Markdown create a markdown-rendering part with a user-specific id.
var MarkdownWithID = tccontent.MarkdownWithID

// Subtitle create a subtitle.
var Subtitle = tccontent.Subtitle

// SubtitleWithID create a subtitle component with a user specific id.
var SubtitleWithID = tccontent.SubtitleWithID

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

// LatexWithID create a latex component with a user specific id.
var LatexWithID = tccontent.LatexWithID
