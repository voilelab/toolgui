# State Cache

State-level cache stores data specific to the current view or "page" displayed to the user.
This data is lost when the user navigates away from the page or refreshes it.

## Example of variable

Here we provide a state-level TODO App example.
It stores the list of todos in the state, so it survives a rerun:

```go
type TODOItem struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

type TODOList struct {
	Items  []TODOItem `json:"items"`
	LastID int        `json:"last_id"`
}

func (t *TODOList) Add(text string) {
	t.LastID++
	t.Items = append(t.Items, TODOItem{ID: t.LastID, Text: text})
}

func Main(p *tgframe.Params) error {
	tgcomp.Title(p.Main, "Example for Todo App")

	todoList := p.State.Default("todoList", &TODOList{}).(*TODOList)

	inp := tgcomp.Textbox(p.State, p.Main, "Add todo")
	if tgcomp.Button(p.State, p.Main, "Add") && inp != "" {
		todoList.Add(inp)
	}

	for i, item := range todoList.Items {
		// The ID ties the checkbox to the item instead of to its position,
		// so its value follows the item when the list changes.
		todoList.Items[i].Done = tgcomp.CheckboxWithConf(p.State, p.Main, item.Text,
			&tgcomp.CheckboxConf{
				ID: fmt.Sprintf("todo_%d", item.ID),
			})
	}

	return nil
}
```

Two things the state makes easy to get wrong:

* A component is sent to the client as soon as it's created, so anything that
  changes the list has to run *before* the list renders. Removing an item after
  the loop leaves it on screen until the next run.
* A component's value is only known once it's rendered. Writing it back to the
  state, as `Done` above, is what lets the next run act on it before rendering.

[`cmd/toolgui-todo`](https://github.com/voilelab/toolgui/blob/dev/cmd/toolgui-todo/main.go)
is the runnable version of this example, with removal on top of it:

```shell
task run_todo
```

## Example of function

Here we provide a state-level cache for a function.

```go
// getFiles unarchive cbz file and return list of file names,
// since unarchive is time-consuming, we use state-level cache to store the result
func getFiles(p *tgframe.Params, f *tcinput.FileObject) ([]string, error) {
	key := fmt.Sprintf("%s_%s_%x", f.Name, f.Type, md5.Sum(f.Bytes))

	v := p.State.GetFuncCache(key)
	if v != nil {
		slog.Debug("cache found")
		return v.([]string), nil
	}

	buf := bytes.NewReader(f.Bytes)

	cbzFp, err := zip.NewReader(buf, buf.Size())
	if err != nil {
		// don't store nil to cache
		return nil, err
	}

	ret := []string{}
	for _, f := range cbzFp.File {
		ret = append(ret, f.Name)
	}

	// ok, we can store to cache
	p.State.SetFuncCache(key, ret)
	return ret, nil
}

func FuncCachePage(p *tgframe.Params) error {
	cbzfile := tgcomp.Fileupload(p.State, p.Sidebar, "CBZ File", "application/x-cbz")

	if cbzfile == nil {
		return nil
	}

	files, err := getFiles(p, cbzfile)
	if err != nil {
		return err
	}

	for i, f := range files {
		tgcomp.Text(p.Main, fmt.Sprintf("%d: %s", i, f))
	}

	return nil
}
```
