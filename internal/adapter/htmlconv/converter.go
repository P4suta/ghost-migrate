package htmlconv

import (
	"bytes"
	"fmt"
	"strings"

	"ghost-migrate/internal/port"

	"golang.org/x/net/html"
)

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) Convert(htmlStr string, rewriter port.URLRewriter) (string, error) {
	if htmlStr == "" {
		return "", nil
	}

	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	ctx := &convertContext{
		buf:       &bytes.Buffer{},
		listDepth: 0,
		rewriter:  rewriter,
	}

	renderChildren(doc, ctx)

	result := ctx.buf.String()
	result = normalizeWhitespace(result)
	return strings.TrimSpace(result) + "\n", nil
}

type convertContext struct {
	buf       *bytes.Buffer
	listDepth int
	rewriter  port.URLRewriter
}

func (ctx *convertContext) rewriteURL(url string) string {
	if ctx.rewriter == nil {
		return url
	}
	return ctx.rewriter(url)
}

func renderNode(n *html.Node, ctx *convertContext) {
	switch n.Type {
	case html.TextNode:
		ctx.buf.WriteString(n.Data)
		return
	case html.ElementNode:
		if isGhostCard(n) {
			renderGhostCard(n, ctx)
			return
		}
		renderElement(n, ctx)
		return
	default:
		renderChildren(n, ctx)
	}
}

func renderChildren(n *html.Node, ctx *convertContext) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderNode(c, ctx)
	}
}

func childrenText(n *html.Node, ctx *convertContext) string {
	oldBuf := ctx.buf
	ctx.buf = &bytes.Buffer{}
	renderChildren(n, ctx)
	result := ctx.buf.String()
	ctx.buf = oldBuf
	return result
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, class string) bool {
	classes := strings.Fields(getAttr(n, "class"))
	for _, c := range classes {
		if c == class {
			return true
		}
	}
	return false
}

func textContent(n *html.Node) string {
	var buf bytes.Buffer
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return buf.String()
}

func normalizeWhitespace(s string) string {
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return s
}
