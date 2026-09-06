# State Cache

State-level cache stores data specific to the current view or "page" displayed to the user.
This data is lost when the user navigates away from the page or refreshes it.

## Example of variable

Here we provide a state-level TODO App example.
It stores list of todos in the state:

```go
type TODOList struct {
	items []string
}

func (t *TODOList) add(item string) {
	t.items = append(t.items, item)
}

func Main(p *tgframe.Params) error {
	tgcomp.Title(p.Main, "Example for Todo App")

	todoList := p.State.Default("todoList", &TODOList{}).(*TODOList)

	inp := tgcomp.Textbox(p.State, p.Main, "Add todo")
	if tgcomp.Button(p.State, p.Main, "Add") {
		todoList.add(inp)
	}

	for i, todo := range todoList.items {
		tgcomp.TextWithID(p.Main,
			fmt.Sprintf("%d: %s", i, todo),
			fmt.Sprintf("todo_%d", i))
	}

	return nil
}
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
