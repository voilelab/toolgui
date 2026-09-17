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
func Badge(c *tgframe.Container, text string, conf ...*BadgeConf) {
	tccontent.Badge(c, text, conf...)
}

// BadgeConf is the configuration for the Badge component.
type BadgeConf = tccontent.BadgeConf

// Caption show a small dimmed text, for a note next to what it explains.
func Caption(c *tgframe.Container, text string, conf ...*CaptionConf) {
	tccontent.Caption(c, text, conf...)
}

// CaptionConf is the configuration for the Caption component.
type CaptionConf = tccontent.CaptionConf

// Code create a code block with syntax highlight.
func Code(c *tgframe.Container, code string, conf ...*CodeConf) {
	tccontent.Code(c, code, conf...)
}

// CodeConf is the configuration for the Code component.
type CodeConf = tccontent.CodeConf

// Divider create a horizontal line.
func Divider(c *tgframe.Container, conf ...*DividerConf) {
	tccontent.Divider(c, conf...)
}

// DividerConf is the configuration for the Divider component.
type DividerConf = tccontent.DividerConf

// Image show a image.
func Image(c *tgframe.Container, img any, conf ...*ImageConf) {
	tccontent.Image(c, img, conf...)
}

// ImageConf is the configuration for the Image component.
type ImageConf = tccontent.ImageConf

// Link create a link component.
func Link(c *tgframe.Container, text, url string, conf ...*LinkConf) {
	tccontent.Link(c, text, url, conf...)
}

// LinkConf is the configuration for the Link component.
type LinkConf = tccontent.LinkConf

// LinkButton create a link that is drawn as a button.
func LinkButton(
	c *tgframe.Container, text, url string, conf ...*LinkButtonConf) {

	tccontent.LinkButton(c, text, url, conf...)
}

// LinkButtonConf is the configuration for the LinkButton component.
type LinkButtonConf = tccontent.LinkButtonConf

// Markdown render markdown to html.
func Markdown(c *tgframe.Container, markdown string, conf ...*MarkdownConf) {
	tccontent.Markdown(c, markdown, conf...)
}

// MarkdownConf is the configuration for the Markdown component.
type MarkdownConf = tccontent.MarkdownConf

// Metric show a labelled value, with an optional delta under it.
func Metric(c *tgframe.Container, label, value string, conf ...*MetricConf) {
	tccontent.Metric(c, label, value, conf...)
}

// MetricConf is the configuration for the Metric component.
type MetricConf = tccontent.MetricConf

// Subtitle create a subtitle.
func Subtitle(c *tgframe.Container, text string, conf ...*SubtitleConf) {
	tccontent.Subtitle(c, text, conf...)
}

// SubtitleConf is the configuration for the Subtitle component.
type SubtitleConf = tccontent.SubtitleConf

// Text show a text.
func Text(c *tgframe.Container, text string, conf ...*TextConf) {
	tccontent.Text(c, text, conf...)
}

// TextConf is the configuration for the Text component.
type TextConf = tccontent.TextConf

// Title show a title.
func Title(c *tgframe.Container, text string, conf ...*TitleConf) {
	tccontent.Title(c, text, conf...)
}

// TitleConf is the configuration for the Title component.
type TitleConf = tccontent.TitleConf

// Latex create a latex component.
func Latex(c *tgframe.Container, text string, conf ...*LatexConf) {
	tccontent.Latex(c, text, conf...)
}

// LatexConf is the configuration for the Latex component.
type LatexConf = tccontent.LatexConf
