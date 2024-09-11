/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package render

import (
	"html/template"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/fsnotify/fsnotify"
)

func TestHTMLDebug_StartChecker_timer(t *testing.T) {
	render := &HTMLDebug{
		RefreshInterval: time.Second,
		Delims:          Delims{Left: "{[{", Right: "}]}"},
		FuncMap:         template.FuncMap{},
		Files:           []string{"../../../common/testdata/template/index.tmpl"},
	}
	select {
	case <-render.reloadCh:
		t.Fatalf("should not be triggered")
	default:
	}
	render.startChecker()
	select {
	case <-time.After(render.RefreshInterval + 500*time.Millisecond):
		t.Fatalf("should be triggered in 1.5 second")
	case <-render.reloadCh:
		render.reload()
	}
}

func TestHTMLDebug_StartChecker_fs_watcher(t *testing.T) {
	f, _ := ioutil.TempFile("./", "test.tmpl")
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	render := &HTMLDebug{Files: []string{f.Name()}}
	select {
	case <-render.reloadCh:
		t.Fatalf("should not be triggered")
	default:
	}
	render.startChecker()
	f.Write([]byte("hello"))
	f.Sync()
	select {
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("should be triggered immediately")
	case <-render.reloadCh:
	}
	select {
	case <-render.reloadCh:
		t.Fatalf("should not be triggered")
	default:
	}
}

func TestRenderHTML(t *testing.T) {
	resp := &protocol.Response{}

	tmpl := template.Must(template.New("").
		Delims("{[{", "}]}").
		Funcs(template.FuncMap{}).
		ParseFiles("../../../common/testdata/template/index.tmpl"))

	r := &HTMLProduction{Template: tmpl}

	html := r.Instance("index.tmpl", utils.H{
		"title": "Main website",
	})

	err := r.Close()
	assert.Nil(t, err)

	html.WriteContentType(resp)
	assert.DeepEqual(t, []byte("text/html; charset=utf-8"), resp.Header.Peek("Content-Type"))

	err = html.Render(resp)

	assert.Nil(t, err)
	assert.DeepEqual(t, []byte("text/html; charset=utf-8"), resp.Header.Peek("Content-Type"))
	assert.DeepEqual(t, []byte("<html><h1>Main website</h1></html>"), resp.Body())

	respDebug := &protocol.Response{}

	rDebug := &HTMLDebug{
		Template: tmpl,
		Delims:   Delims{Left: "{[{", Right: "}]}"},
		FuncMap:  template.FuncMap{},
		Files:    []string{"../../../common/testdata/template/index.tmpl"},
	}

	htmlDebug := rDebug.Instance("index.tmpl", utils.H{
		"title": "Main website",
	})

	err = rDebug.Close()
	assert.Nil(t, err)

	htmlDebug.WriteContentType(respDebug)
	assert.DeepEqual(t, []byte("text/html; charset=utf-8"), respDebug.Header.Peek("Content-Type"))

	err = htmlDebug.Render(respDebug)

	assert.Nil(t, err)
	assert.DeepEqual(t, []byte("text/html; charset=utf-8"), respDebug.Header.Peek("Content-Type"))
	assert.DeepEqual(t, []byte("<html><h1>Main website</h1></html>"), respDebug.Body())
}

func TestHTMLProduction_Instance(t *testing.T) {
	tmpl := template.Must(template.New("").Parse("<h1>{{.Title}}</h1>"))
	r := &HTMLProduction{Template: tmpl}

	html := r.Instance("test", map[string]interface{}{"Title": "Test Title"})

	assert.NotNil(t, html)
	assert.DeepEqual(t, tmpl, html.(HTML).Template)
	assert.DeepEqual(t, "test", html.(HTML).Name)
	assert.DeepEqual(t, map[string]interface{}{"Title": "Test Title"}, html.(HTML).Data)
}

func TestHTMLDebug_Instance(t *testing.T) {
	tmpl := template.Must(template.New("").Parse("<h1>{{.Title}}</h1>"))
	r := &HTMLDebug{
		Template: tmpl,
		Files:    []string{"test.tmpl"},
	}

	html := r.Instance("test", map[string]interface{}{"Title": "Test Title"})

	assert.NotNil(t, html)
	assert.DeepEqual(t, tmpl, html.(HTML).Template)
	assert.DeepEqual(t, "test", html.(HTML).Name)
	assert.DeepEqual(t, map[string]interface{}{"Title": "Test Title"}, html.(HTML).Data)
}

func TestHTMLDebug_Close(t *testing.T) {
	r := &HTMLDebug{}
	err := r.Close()
	assert.Nil(t, err)

	watcher, _ := fsnotify.NewWatcher()
	r.watcher = watcher
	err = r.Close()
	assert.Nil(t, err)
}

func TestHTMLDebug_reload(t *testing.T) {
	tmpl := template.Must(template.New("").Parse("<h1>{{.Title}}</h1>"))
	r := &HTMLDebug{
		Template: tmpl,
		Files:    []string{"test.tmpl"},
		Delims:   Delims{Left: "{{", Right: "}}"},
		FuncMap:  template.FuncMap{},
	}

	r.reload()

	assert.NotNil(t, r.Template)
	assert.NotEqual(t, tmpl, r.Template)
}

func TestHTMLDebug_startChecker(t *testing.T) {
	r := &HTMLDebug{
		RefreshInterval: 100 * time.Millisecond,
		Files:           []string{"test.tmpl"},
	}

	r.startChecker()

	assert.NotNil(t, r.reloadCh)

	select {
	case <-r.reloadCh:
		// Reload triggered
	case <-time.After(150 * time.Millisecond):
		t.Fatal("Reload not triggered within expected time")
	}
}
