package templatecollection

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"
)

type Collection interface {
	ExecuteTemplate(wr io.Writer, name string, data interface{}) error
}

var ErrTemplateNotFound = fmt.Errorf("template not found")

type Cached struct {
	l sync.RWMutex
	m map[string]*template.Template
}

func NewCached(fileSystem fs.FS, funcs template.FuncMap) (Collection, error) {
	pageFiles, err := fs.Glob(fileSystem, "**/page_*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("templatecollection.NewCached: could not get page template names: %w", err)
	}

	c := Cached{m: make(map[string]*template.Template)}

	for _, pageFile := range pageFiles {
		name := strings.TrimSuffix(path.Base(pageFile), ".gohtml")

		tpl := template.New(name)
		if funcs != nil {
			tpl = tpl.Funcs(funcs)
		}

		fileNames := []string{}

		for _, pattern := range expandGlobs([]string{pageFile, "layout.gohtml", "shared_*.gohtml"}) {
			names, err := fs.Glob(fileSystem, pattern)
			if err != nil {
				return nil, fmt.Errorf("templatecollection.NewCached: could not get names for pattern %q: %w", pattern, err)
			}

			fileNames = append(fileNames, names...)
		}

		tpl, err := tpl.ParseFS(fileSystem, fileNames...)
		if err != nil {
			return nil, fmt.Errorf("templatecollection.NewCached: could not construct template: %w", err)
		}

		c.m[name] = tpl
	}

	return &c, nil
}

func (c *Cached) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	c.l.RLock()
	defer c.l.RUnlock()

	tpl, ok := c.m[name]
	if !ok {
		return fmt.Errorf("templatecollection.Cached.ExecuteTemplate: %w", ErrTemplateNotFound)
	}

	if err := tpl.ExecuteTemplate(wr, name, data); err != nil {
		return fmt.Errorf("templatecollection.Cached.ExecuteTemplate: %w", err)
	}

	return nil
}

type Live struct {
	fs fs.FS
	m  template.FuncMap
}

func NewLive(fileSystem fs.FS, funcs template.FuncMap) (Collection, error) {
	return &Live{fs: fileSystem, m: funcs}, nil
}

func (l *Live) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	tpl := template.New(name)
	if l.m != nil {
		tpl = tpl.Funcs(l.m)
	}

	fileNames := []string{}

	for _, pattern := range expandGlobs([]string{name + ".gohtml", "layout.gohtml", "shared_*.gohtml"}) {
		names, err := fs.Glob(l.fs, pattern)
		if err != nil {
			return fmt.Errorf("templatecollection.Live.ExecuteTemplate: could not get names for pattern %q: %w", pattern, err)
		}

		fileNames = append(fileNames, names...)
	}

	tpl, err := tpl.ParseFS(l.fs, fileNames...)
	if err != nil {
		return fmt.Errorf("templatecollection.Live.ExecuteTemplate: could not construct template: %w", err)
	}

	if err := tpl.ExecuteTemplate(wr, name, data); err != nil {
		return fmt.Errorf("templatecollection.Live.ExecuteTemplate: %w", err)
	}

	return nil
}

func expandGlobs(a []string) []string {
	var r []string

	for _, e := range a {
		r = append(r, e, "**/"+e)
	}

	return r
}
