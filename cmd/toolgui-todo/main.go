package main

import (
	"fmt"
	"log"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgexec"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// TODOItem is one todo. ID identifies it for its whole life, so a checkbox
// stays tied to the item instead of to its position in the list.
type TODOItem struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// TODOList is the app data. It lives in the state, so it survives a rerun.
type TODOList struct {
	Items  []TODOItem `json:"items"`
	LastID int        `json:"last_id"`
}

func (t *TODOList) Add(text string) {
	t.LastID++
	t.Items = append(t.Items, TODOItem{ID: t.LastID, Text: text})
}

// RemoveDone drops every done item.
func (t *TODOList) RemoveDone() {
	rest := make([]TODOItem, 0, len(t.Items))
	for _, item := range t.Items {
		if !item.Done {
			rest = append(rest, item)
		}
	}

	t.Items = rest
}

// DoneTexts returns the text of every done item.
func (t *TODOList) DoneTexts() []string {
	texts := []string{}
	for _, item := range t.Items {
		if item.Done {
			texts = append(texts, item.Text)
		}
	}

	return texts
}

func Main(p *tgframe.Params) error {
	tgcomp.Title(p.Main, "Example for Todo App / State")

	todoList := p.State.Default("todoList", &TODOList{}).(*TODOList)

	col1, col2 := tgcomp.EqColumn2(p.Main, "divided")

	tgcomp.Text(col1, "App")

	inp := tgcomp.Textbox(p.State, col1, "Add todo")
	if tgcomp.Button(p.State, col1, "Add") && inp != "" {
		todoList.Add(inp)
	}

	// Both mutations happen before the list renders. A component is sent to
	// the client as soon as it's created, so a removed item would stay on
	// screen until the next run if we removed it after the loop.
	if tgcomp.Button(p.State, col1, "Remove done") {
		todoList.RemoveDone()
	}

	for i, item := range todoList.Items {
		// Checkbox returns its own value, and it's written back to the state
		// so the next run knows which items are done before rendering them.
		todoList.Items[i].Done = tgcomp.CheckboxWithConf(p.State, col1, item.Text,
			&tgcomp.CheckboxConf{
				ID: fmt.Sprintf("todo_%d", item.ID),
			})
	}

	tgcomp.Text(col2, "Selected State")
	tgcomp.JSON(col2, todoList.DoneTexts())

	return nil
}

func main() {
	app := tgframe.NewApp()
	app.AddPage("main", "Main", Main)

	e := tgexec.NewWebExecutor(app)
	log.Println("Starting service...")
	err := e.StartService(":3000")
	if err != nil {
		log.Println(err)
	}
}
