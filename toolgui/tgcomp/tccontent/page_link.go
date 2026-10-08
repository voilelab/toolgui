package tccontent

import (
	"net/url"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

var _ tgframe.Component = &pageLinkComponent{}
var pageLinkComponentName = "page_link_component"

type pageLinkComponent struct {
	*tgframe.BaseComponent
	Text  string     `json:"text"`
	Page  string     `json:"page"`
	Query url.Values `json:"query"`
}

func newPageLinkComponent(
	text, page string, query url.Values) *pageLinkComponent {

	// Never nil, so the client always reads an object.
	if query == nil {
		query = url.Values{}
	}

	return &pageLinkComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: pageLinkComponentName,
		},
		Text:  text,
		Page:  page,
		Query: query,
	}
}

// PageLinkConf is the configuration for the PageLink component.
type PageLinkConf struct {
	tgframe.Base
}

// PageLink create a link to page of this app, opened with query as its
// [tgframe.Params.Query]. It is a real link, so middle-click and copying it
// work. It takes a page name, not a URL, so it cannot point off the app.
// An unknown page fails the component.
func PageLink(c *tgframe.Container, text, page string,
	query url.Values, conf ...*PageLinkConf) {

	cf := tgframe.OneConf("PageLink", conf)

	if !c.HasPage(page) {
		c.Fail(tgutil.Errorf("%w: `%s`", tgframe.ErrPageNotFound, page))
		return
	}

	comp := newPageLinkComponent(text, page, query)
	tgframe.SetConfIDIn(c, comp, cf)
	c.AddComponent(comp)
}
