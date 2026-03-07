package htmlconv

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func renderElement(n *html.Node, ctx *convertContext) {
	switch n.Data {
	case "p":
		text := childrenText(n, ctx)
		if strings.TrimSpace(text) != "" {
			fmt.Fprintf(ctx.buf, "\n%s\n", text)
		}

	case "h1":
		fmt.Fprintf(ctx.buf, "\n# %s\n", childrenText(n, ctx))
	case "h2":
		fmt.Fprintf(ctx.buf, "\n## %s\n", childrenText(n, ctx))
	case "h3":
		fmt.Fprintf(ctx.buf, "\n### %s\n", childrenText(n, ctx))
	case "h4":
		fmt.Fprintf(ctx.buf, "\n#### %s\n", childrenText(n, ctx))
	case "h5":
		fmt.Fprintf(ctx.buf, "\n##### %s\n", childrenText(n, ctx))
	case "h6":
		fmt.Fprintf(ctx.buf, "\n###### %s\n", childrenText(n, ctx))

	case "blockquote":
		lines := strings.Split(strings.TrimSpace(childrenText(n, ctx)), "\n")
		ctx.buf.WriteString("\n")
		for _, line := range lines {
			fmt.Fprintf(ctx.buf, "> %s\n", line)
		}

	case "hr":
		ctx.buf.WriteString("\n---\n")

	case "br":
		ctx.buf.WriteString("\n")

	case "pre":
		renderCodeBlock(n, ctx)

	case "ul":
		renderList(n, ctx, false)
	case "ol":
		renderList(n, ctx, true)
	case "li":
		renderChildren(n, ctx)

	case "strong", "b":
		fmt.Fprintf(ctx.buf, "**%s**", childrenText(n, ctx))
	case "em", "i":
		fmt.Fprintf(ctx.buf, "*%s*", childrenText(n, ctx))
	case "code":
		fmt.Fprintf(ctx.buf, "`%s`", childrenText(n, ctx))
	case "del", "s", "strike":
		fmt.Fprintf(ctx.buf, "~~%s~~", childrenText(n, ctx))
	case "u":
		renderChildren(n, ctx)
	case "mark":
		fmt.Fprintf(ctx.buf, "==%s==", childrenText(n, ctx))
	case "sup":
		fmt.Fprintf(ctx.buf, "^%s^", childrenText(n, ctx))
	case "sub":
		fmt.Fprintf(ctx.buf, "~%s~", childrenText(n, ctx))

	case "a":
		href := ctx.rewriteURL(getAttr(n, "href"))
		text := childrenText(n, ctx)
		if href != "" {
			fmt.Fprintf(ctx.buf, "[%s](%s)", text, href)
		} else {
			ctx.buf.WriteString(text)
		}

	case "img":
		src := ctx.rewriteURL(getAttr(n, "src"))
		alt := getAttr(n, "alt")
		fmt.Fprintf(ctx.buf, "![%s](%s)", alt, src)

	case "figure":
		if isGhostCard(n) {
			renderGhostCard(n, ctx)
		} else {
			renderFigure(n, ctx)
		}

	case "iframe":
		src := ctx.rewriteURL(getAttr(n, "src"))
		if src != "" {
			fmt.Fprintf(ctx.buf, "\n%s\n", src)
		}

	case "table":
		renderTable(n, ctx)

	default:
		renderChildren(n, ctx)
	}
}

func renderCodeBlock(n *html.Node, ctx *convertContext) {
	lang := ""
	var codeNode *html.Node

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "code" {
			codeNode = c
			class := getAttr(c, "class")
			if strings.HasPrefix(class, "language-") {
				lang = strings.TrimPrefix(class, "language-")
			}
			break
		}
	}

	var code string
	if codeNode != nil {
		code = textContent(codeNode)
	} else {
		code = textContent(n)
	}

	fmt.Fprintf(ctx.buf, "\n```%s\n%s\n```\n", lang, code)
}

func renderList(n *html.Node, ctx *convertContext, ordered bool) {
	ctx.buf.WriteString("\n")
	index := 1
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.Data != "li" {
			continue
		}

		indent := strings.Repeat("  ", ctx.listDepth)
		if ordered {
			fmt.Fprintf(ctx.buf, "%s%d. ", indent, index)
			index++
		} else {
			fmt.Fprintf(ctx.buf, "%s- ", indent)
		}

		for child := c.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && (child.Data == "ul" || child.Data == "ol") {
				ctx.listDepth++
				renderList(child, ctx, child.Data == "ol")
				ctx.listDepth--
			} else {
				renderNode(child, ctx)
			}
		}

		if !strings.HasSuffix(ctx.buf.String(), "\n") {
			ctx.buf.WriteString("\n")
		}
	}
}

func renderFigure(n *html.Node, ctx *convertContext) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			switch c.Data {
			case "img":
				renderNode(c, ctx)
			case "figcaption":
				text := strings.TrimSpace(childrenText(c, ctx))
				if text != "" {
					fmt.Fprintf(ctx.buf, "\n*%s*", text)
				}
			case "iframe":
				renderNode(c, ctx)
			default:
				renderNode(c, ctx)
			}
		}
	}
	ctx.buf.WriteString("\n")
}

func renderTable(n *html.Node, ctx *convertContext) {
	rows := collectTableRows(n)
	if len(rows) == 0 {
		return
	}

	ctx.buf.WriteString("\n")

	ctx.buf.WriteString("|")
	for _, cell := range rows[0] {
		fmt.Fprintf(ctx.buf, " %s |", cell)
	}
	ctx.buf.WriteString("\n|")
	for range rows[0] {
		ctx.buf.WriteString(" --- |")
	}
	ctx.buf.WriteString("\n")

	for _, row := range rows[1:] {
		ctx.buf.WriteString("|")
		for _, cell := range row {
			fmt.Fprintf(ctx.buf, " %s |", cell)
		}
		ctx.buf.WriteString("\n")
	}
}

func collectTableRows(n *html.Node) [][]string {
	var rows [][]string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var cells []string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
					cells = append(cells, strings.TrimSpace(textContent(c)))
				}
			}
			if len(cells) > 0 {
				rows = append(rows, cells)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return rows
}
