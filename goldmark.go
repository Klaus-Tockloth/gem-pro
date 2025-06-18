package main

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// TargetBlankHTMLRenderer is a struct for the custom link renderer.
type TargetBlankHTMLRenderer struct {
	html.Config
}

// NewTargetBlankHTMLRenderer creates a new renderer.
func NewTargetBlankHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &TargetBlankHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// Render processes the link node.
func (r *TargetBlankHTMLRenderer) Render(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		_, _ = w.WriteString("<a href=\"")
		if r.Unsafe || !html.IsDangerousURL(n.Destination) {
			_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		}
		_, _ = w.WriteString("\"")
		// Add target="_blank" and rel="noopener noreferrer"
		_, _ = w.WriteString(` target="_blank" rel="noopener noreferrer"`)
		if n.Title != nil {
			_, _ = w.WriteString(` title="`)
			r.Writer.Write(w, n.Title)
			_ = w.WriteByte('"')
		}
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}

// RegisterFuncs registers the render function for Link nodes.
func (r *TargetBlankHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindLink, r.Render)
	reg.Register(ast.KindAutoLink, r.Render) // Optional: also handle Auto-Links
}

// TargetBlankExtension is a struct for the extension.
type TargetBlankExtension struct{}

// Extend adds the custom renderer.
func (e *TargetBlankExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewTargetBlankHTMLRenderer(), 1), // Priority 1 overrides the default
	))
}
