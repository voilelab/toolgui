package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgexec"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

type TODOList struct {
	items []string
}

func (t *TODOList) add(item string) {
	t.items = append(t.items, item)
}

func Main(p *tgframe.Params) error {
	tgcomp.Title(p.Main, "Example for Todo App / State")

	todoList := p.State.Default("todoList", &TODOList{}).(*TODOList)

	col1, col2 := tgcomp.EqColumn2(p.Main, "divided")

	tgcomp.Text(col1, "App")

	inp := tgcomp.Textbox(p.State, col1, "Add todo")
	if tgcomp.Button(p.State, col1, "Add") {
		todoList.add(inp)
	}

	selectItems := []string{}
	for i, todo := range todoList.items {
		// FIXME: provide a way to get the id of the checkbox component
		checkboxID := fmt.Sprintf("checkbox_component_todo_%d", i)
		if p.State.GetBool(checkboxID) {
			selectItems = append(selectItems, todo)
		}
	}

	tgcomp.Text(col2, "Selected State")
	tgcomp.JSON(col2, selectItems)

	for i, todo := range todoList.items {
		tgcomp.CheckboxWithConf(p.State, col1, fmt.Sprintf("%d: %s", i, todo), &tgcomp.CheckboxConf{
			ID: fmt.Sprintf("todo_%d", i),
		})
	}

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
