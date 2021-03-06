package formatting

import (
	"bytes"
	"io"
	"sync"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

type MkdHtmlRenderer struct {
	blackfriday.Renderer
}

func (h *MkdHtmlRenderer) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	language := LanguageNamed(lang)
	if language == nil {
		h.Renderer.BlockCode(out, text, lang)
		return
	}
	r := bytes.NewReader(text)
	rendered, err := FormatStream(r, language)
	if err == nil {
		// TODO(DH): remove html from formatters subpackage.
		out.WriteString(`<div class="code code-` + language.DisplayStyle + `">` + rendered + `</div>`)
	} else {
		out.WriteString(`<div class="well well-error"><i class="icon icon-warning"></i> <strong>Code block failed to render.</strong><br></div>`)
	}
}

func NewMkdHtmlRenderer() *MkdHtmlRenderer {
	return &MkdHtmlRenderer{blackfriday.HtmlRenderer(blackfriday.HTML_SAFELINK|
		blackfriday.HTML_NOFOLLOW_LINKS, "", "")}
}

var mkdHtmlRenderer blackfriday.Renderer
var sanitationPolicy *bluemonday.Policy
var markdownOnce sync.Once

func markdownFormatter(formatter *Formatter, stream io.Reader, args ...string) (string, error) {
	markdownOnce.Do(func() {
		// one-time init
		mkdHtmlRenderer = NewMkdHtmlRenderer()
		sanitationPolicy = bluemonday.UGCPolicy()
		sanitationPolicy.AllowAttrs("class").OnElements("div", "i", "span")
	})

	buf := &bytes.Buffer{}
	io.Copy(buf, stream)
	md := blackfriday.Markdown(buf.Bytes(), mkdHtmlRenderer,
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS|
			blackfriday.EXTENSION_TABLES|
			blackfriday.EXTENSION_AUTOLINK|
			blackfriday.EXTENSION_FENCED_CODE|
			blackfriday.EXTENSION_HEADER_IDS|
			blackfriday.EXTENSION_LAX_HTML_BLOCKS)
	return sanitationPolicy.Sanitize(string(md)), nil
}
