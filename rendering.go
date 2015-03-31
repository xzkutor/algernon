package main

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/eknkc/amber"
	"github.com/russross/blackfriday"
	"github.com/yosssi/gcss"
	"github.com/yuin/gopher-lua"
	"io"
	"net/http"
	"strings"
)

// Expose functions that are related to rendering text, to the given Lua state
func exportRenderFunctions(w http.ResponseWriter, req *http.Request, L *lua.LState) {

	// Output Markdown as HTML
	L.SetGlobal("mprint", L.NewFunction(func(L *lua.LState) int {
		buf := arguments2buffer(L)
		// Convert the buffer to markdown and return the translated string
		w.Write(blackfriday.MarkdownCommon([]byte(buf.String())))
		return 0 // number of results
	}))

	// TODO: Add two functions. One to compile amber templates and
	// store the result by filename and one to render data by using
	// compiled templates.

	// Output text as rendered amber.
	// TODO: Add caching, compilation and reuse
	L.SetGlobal("aprint", L.NewFunction(func(L *lua.LState) int {
		buf := arguments2buffer(L)

		// Use the buffer as a template.
		// Options are "Pretty printing, but without line numbers."
		tpl, err := amber.Compile(buf.String(), amber.Options{true, false})
		if err != nil {
			if DEBUG_MODE {
				// TODO: Show where in the source code things went wrong. Make it prettier.
				fmt.Fprint(w, "Could not compile Amber template:\n\t"+err.Error()+"\n\n"+buf.String())
			} else {
				log.Error("Could not compile Amber template:")
				log.Error(err)
				log.Error("\n" + buf.String())
			}
			return 0 // number of results
		}
		//somedata := map[string]string{"": ""}
		tpl.Execute(w, nil)
		return 0 // number of results
	}))

	// Output text as rendered GCSS
	// TODO: Add caching, compilation and reuse
	L.SetGlobal("gprint", L.NewFunction(func(L *lua.LState) int {
		buf := arguments2buffer(L)
		// Transform GCSS to CSS and output the result.
		// Ignoring the number of bytes written.
		if _, err := gcss.Compile(w, bytes.NewReader(buf.Bytes())); err != nil {
			if DEBUG_MODE {
				// TODO: Show where in the source code things went wrong. Make it prettier.
				fmt.Fprint(w, "Could not compile GCSS:\n\t"+err.Error()+"\n\n"+buf.String())
			} else {
				log.Error("Could not compile GCSS:")
				log.Error(err)
				log.Error("\n" + buf.String())
			}
			//return 0 // number of results
		}
		return 0 // number of results
	}))

}

// Write the given source bytes as markdown wrapped in HTML to a writer, with a title
func markdownPage(w io.Writer, b []byte, title string) {

	// Convert from Markdown to HTML
	htmlbody := string(blackfriday.MarkdownCommon(b))

	// TODO: Check if handling "# title <tags" on the first line is valid
	// Markdown or not. Submit a patch to blackfriday if it is.

	h1title := ""
	if strings.HasPrefix(htmlbody, "<p>#") {
		fields := strings.Split(htmlbody, "<")
		if len(fields) > 2 {
			h1title = fields[1][2:]
			htmlbody = htmlbody[len("<p>"+h1title):]
			if strings.HasPrefix(h1title, "#") {
				h1title = h1title[1:]
			}
		}
	}

	htmlbytes := []byte("<!doctype html><html><head><title>" + title + "</title><style>" + style + "</style><head><body><h1>" + h1title + "</h1>" + htmlbody + "</body></html>")
	w.Write(htmlbytes)
}

// Write the given source bytes as Amber converted to HTML, to a writer.
func amberPage(w io.Writer, b []byte, title string) {
	ambertext := string(b)
	tpl, err := amber.Compile(ambertext, amber.Options{true, false})
	if err != nil {
		if DEBUG_MODE {
			// TODO: Show where in the source code things went wrong. Make it prettier.
			fmt.Fprint(w, "Could not compile Amber template:\n\t"+err.Error()+"\n\n"+ambertext)
		} else {
			log.Error("Could not compile Amber template:")
			log.Error(err.Error())
			log.Error("\n" + ambertext)
		}
		return

	}
	//somedata := map[string]string{"": ""}
	var buf bytes.Buffer
	//bufWriter := bufio.NewWriter(&buf)
	if err := tpl.Execute(&buf, "MISSING DATA"); err != nil {
		if DEBUG_MODE {
			// TODO: Make it prettier.
			fmt.Fprint(w, "Could not execute Amber template:\n\t"+err.Error())
		} else {
			log.Error("Could not execute Amber template:")
			log.Error(err)
		}
		return
	}
	buf.WriteTo(w)
}

// Write the given source bytes as GCSS converted to CSS, to a writer.
func gcssPage(w io.Writer, b []byte, title string) {
	if _, err := gcss.Compile(w, bytes.NewReader(b)); err != nil {
		if DEBUG_MODE {
			// TODO: Show where in the source code things went wrong. Make it prettier.
			fmt.Fprint(w, "Could not compile GCSS:\n\t"+err.Error()+"\n\n"+string(b))
		} else {
			log.Error("Could not compile GCSS:")
			log.Error(err)
			log.Error("\n" + string(b))
		}
		return
	}
}
