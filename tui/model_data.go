package tui

import (
	awsqueries "codeplumber/aws"
)

// DataCache is a struct to hold the data cache
type DataCache struct {
	pipelines  map[string]awsqueries.Pipeline
	codebuilds map[string]awsqueries.CodebuildData
}

// uiData is a struct to hold UI data shared accoss all the views
type uiData struct {
	selection     chan tuiMsg
	statusCmd     string
	dataCache     DataCache
	viewIdx       int // TODO: Should be moved to Model
	views         []string
	refreshing    bool
	inputFocused  bool
	initialized   bool
	path          []string
	width, height int
	help          bool
}

func (c *uiData) startSpinner() {
	c.refreshing = true
	c.statusCmd = "spinner"
}

func (c *uiData) stopSpinner() {
	c.refreshing = false
	c.statusCmd = ""
}

func (c *uiData) currentView() string {
	return c.views[c.viewIdx]
}

func (c *uiData) updatPath(name string) {
	if c.path[c.viewIdx] != name {
		c.path[c.viewIdx] = name
	}
}

func (c *uiData) errorMsg(src, msg string) {
	c.selection <- tuiMsg{
		class: errorMsg,
		src:   src,
		data:  msg,
	}
}

func (c *uiData) confirm(src, msg string, ref interface{}) {
	c.selection <- tuiMsg{
		class:     input,
		id:        "confirm",
		src:       src,
		data:      msg,
		reference: ref,
	}
}

func (c *uiData) search(src string) {
	c.selection <- tuiMsg{
		class:     input,
		id:        "search",
		src:       src,
		data:      "",
		reference: "",
	}
}

func (c *uiData) requestInput(src, id string, data, ref interface{}) {
	c.selection <- tuiMsg{
		class:     input,
		src:       src,
		id:        id,
		data:      data,
		reference: ref,
	}
}

func (c *uiData) responseInput(src, id string, data, ref interface{}, trigger bool) {
	c.selection <- tuiMsg{
		class:     response,
		src:       src,
		id:        id,
		data:      data,
		reference: ref,
		trigger:   trigger,
	}
}

func (c *uiData) updateView(id string, data interface{}) {
	c.selection <- tuiMsg{
		class: viewUpdate,
		id:    id,
		data:  data,
	}
}

func (c *uiData) changeView(src, dst string, selection interface{}) {
	c.selection <- tuiMsg{
		class: viewChange,
		src:   src,
		id:    dst,
		data:  selection,
	}
}

func (c *uiData) previousView() {
	c.selection <- tuiMsg{
		class: previous,
	}
}

func (c *uiData) previousViewWithRefresh() {
	c.selection <- tuiMsg{
		class:   previous,
		trigger: true,
	}
}
