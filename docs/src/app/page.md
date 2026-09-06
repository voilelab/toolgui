# Page

![](layout.png)

## Parameters

We can config the page's:

1. Name: Or ID. The name of the page should be unique.
In the web GUI provider, the name of a page will be used as the path of the page.

2. Title: In the web GUI provider,
the title of a page will be used as the title of the page and the text of its link in the side nav.

3. Emoji: Optional. The emoji will be used as an icon in the side nav and browser favicon.

The config type in package is:

```go
type PageConfig struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Emoji string `json:"emoji"`
}
```

## Page Function

The function signature of **Page Function** is defined as:

```go
type RunFunc func(p *Params) error
```

Where `Params` contains these parameters to operate the page:

```go
type Params struct {
	State   *State
	Main    *Container
	Sidebar *Container
}
```

`Main` and `Sidebar` is the root container component of the main and sidebar part
shown in the layout image.

We can show a text in the Main container by:

```go
tgcomp.Text(p.Main, "Hello")
```

The `State` in `Params` provided for

1. The component that need to pass state. For example: the checked state of checkbox.
2. The state that user need to store. For example: The todo items in the Todo App.

## Example for adding a page

* No emoji icon

```go
app.AddPage("index", "Index", Main)
```

* With a emoji icon

```go
app.AddPageByConfig(&tgframe.PageConfig{
	Name:  "page2",
	Title: "Page2",
	Emoji: "🔄",
}, Page2)
```

## Page name in the URL

By default the web executor serves each page at its own path, `/{name}`, and
redirects `/` to the first page added.

An app served from a static path — or opened from a file — cannot rely on the
server routing those paths. `SetHashPageNameMode` moves the page name into the
URL fragment, `/#/{name}`, so every page is served from `/`:

```go
app.SetHashPageNameMode(true)
```

With no fragment, the first page added is shown.
