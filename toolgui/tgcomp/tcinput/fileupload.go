package tcinput

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

var _ tgframe.Component = &fileuploadComponent{}
var fileuploadComponentName = "fileupload_component"

type fileuploadComponent struct {
	*tgframe.BaseComponent
	Label    string `json:"label"`
	Accept   string `json:"accept"`
	Disabled bool   `json:"disabled"`
}

func newFileuploadComponent(label, accept string) *fileuploadComponent {
	return &fileuploadComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: fileuploadComponentName,
			ID:   tcutil.NormalID(fileuploadComponentName, label),
		},
		Label:  label,
		Accept: accept,
	}
}

// FileObject is the object that is returned when a file is uploaded.
//
// The content stays on disk. Read it with [FileObject.Open] to work through a
// stream, or [FileObject.Bytes] to take it whole.
type FileObject struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int    `json:"size"`

	file *tgframe.File
}

// Open return a reader over the uploaded content. The caller closes it.
func (f *FileObject) Open() (tgframe.FileReader, error) {
	if f.file == nil {
		return nil, tgutil.NewError("file object has no content")
	}

	return f.file.Open()
}

// Bytes read the whole upload into memory. Prefer [FileObject.Open] for
// anything that can work on a stream: a file only has to fit on disk.
func (f *FileObject) Bytes() ([]byte, error) {
	if f.file == nil {
		return nil, tgutil.NewError("file object has no content")
	}

	return f.file.Bytes()
}

// FileuploadConf is the configuration for the Fileupload component.
//
// There is no Default here, unlike the other inputs. A file input is the one
// control a page cannot fill in on the app user's behalf: the browser refuses
// to have its value set from script, so a default would read back in Go while
// the box on screen stayed empty. A page that wants to start from a file it
// already has should read that file itself rather than ask for one.
type FileuploadConf struct {
	tgframe.Base

	// Disabled is true if the fileupload is disabled.
	Disabled bool
}

// Fileupload create a fileupload and return its selected file.
// Return nil if no file is selected.
func Fileupload(c *tgframe.Container, label, accept string, conf ...*FileuploadConf) *FileObject {
	cf := tgframe.OneConf("Fileupload", conf)

	comp := newFileuploadComponent(label, accept)
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)

	var fileObj *FileObject
	err := c.State.GetObject(comp.ID, &fileObj)
	if err != nil {
		panic(err)
	}

	if fileObj == nil {
		return nil
	}

	// The content is stored under the component, so a second fileupload that
	// takes a file of the same name doesn't take this one's content with it.
	fileObj.file = c.State.GetFile(comp.ID)
	if fileObj.file == nil {
		// The pick reached the state but the upload didn't.
		return nil
	}

	// Size arrives from the browser. What was stored is what a reader will
	// actually get, so that's what the page is told.
	fileObj.Size = int(fileObj.file.Size())

	return fileObj
}
