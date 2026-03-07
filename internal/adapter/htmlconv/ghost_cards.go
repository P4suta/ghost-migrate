package htmlconv

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func isGhostCard(n *html.Node) bool {
	return hasClass(n, "kg-card")
}

func renderGhostCard(n *html.Node, ctx *convertContext) {
	switch {
	case hasClass(n, "kg-image-card"):
		renderImageCard(n, ctx)
	case hasClass(n, "kg-gallery-card"):
		renderGalleryCard(n, ctx)
	case hasClass(n, "kg-code-card"):
		renderChildren(n, ctx)
	case hasClass(n, "kg-bookmark-card"):
		renderBookmarkCard(n, ctx)
	case hasClass(n, "kg-callout-card"):
		renderCalloutCard(n, ctx)
	case hasClass(n, "kg-toggle-card"):
		renderToggleCard(n, ctx)
	case hasClass(n, "kg-embed-card"):
		renderEmbedCard(n, ctx)
	case hasClass(n, "kg-button-card"):
		renderButtonCard(n, ctx)
	case hasClass(n, "kg-product-card"):
		renderProductCard(n, ctx)
	case hasClass(n, "kg-file-card"):
		renderFileCard(n, ctx)
	default:
		renderChildren(n, ctx)
	}
}

func renderImageCard(n *html.Node, ctx *convertContext) {
	var imgSrc, imgAlt, caption string

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "img":
				imgSrc = ctx.rewriteURL(getAttr(node, "src"))
				imgAlt = getAttr(node, "alt")
			case "figcaption":
				caption = strings.TrimSpace(textContent(node))
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	if imgSrc != "" {
		fmt.Fprintf(ctx.buf, "\n![%s](%s)", imgAlt, imgSrc)
		if caption != "" {
			fmt.Fprintf(ctx.buf, "\n*%s*", caption)
		}
		ctx.buf.WriteString("\n")
	}
}

func renderGalleryCard(n *html.Node, ctx *convertContext) {
	ctx.buf.WriteString("\n")
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "img" {
			src := ctx.rewriteURL(getAttr(node, "src"))
			alt := getAttr(node, "alt")
			fmt.Fprintf(ctx.buf, "![%s](%s)\n\n", alt, src)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
}

func renderBookmarkCard(n *html.Node, ctx *convertContext) {
	var href, title, description string

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if node.Data == "a" && hasClass(node, "kg-bookmark-container") {
				href = ctx.rewriteURL(getAttr(node, "href"))
			}
			if hasClass(node, "kg-bookmark-title") {
				title = strings.TrimSpace(textContent(node))
			}
			if hasClass(node, "kg-bookmark-description") {
				description = strings.TrimSpace(textContent(node))
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	if href != "" {
		fmt.Fprintf(ctx.buf, "\n> **[%s](%s)**\n", title, href)
		if description != "" {
			fmt.Fprintf(ctx.buf, "> %s\n", description)
		}
	}
}

func renderCalloutCard(n *html.Node, ctx *convertContext) {
	var emoji, text string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if hasClass(node, "kg-callout-emoji") {
				emoji = strings.TrimSpace(textContent(node))
			}
			if hasClass(node, "kg-callout-text") {
				text = strings.TrimSpace(textContent(node))
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	if emoji != "" {
		fmt.Fprintf(ctx.buf, "\n> %s %s\n", emoji, text)
	} else {
		fmt.Fprintf(ctx.buf, "\n> %s\n", text)
	}
}

func renderToggleCard(n *html.Node, ctx *convertContext) {
	var heading, content string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if hasClass(node, "kg-toggle-heading-text") {
				heading = strings.TrimSpace(textContent(node))
			}
			if hasClass(node, "kg-toggle-content") {
				content = strings.TrimSpace(textContent(node))
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	fmt.Fprintf(ctx.buf, "\n<details>\n<summary>%s</summary>\n\n%s\n\n</details>\n", heading, content)
}

func renderEmbedCard(n *html.Node, ctx *convertContext) {
	var src, caption string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if node.Data == "iframe" {
				src = ctx.rewriteURL(getAttr(node, "src"))
			}
			if node.Data == "figcaption" {
				caption = strings.TrimSpace(textContent(node))
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	if src != "" {
		fmt.Fprintf(ctx.buf, "\n%s\n", src)
		if caption != "" {
			fmt.Fprintf(ctx.buf, "*%s*\n", caption)
		}
	}
}

func renderButtonCard(n *html.Node, ctx *convertContext) {
	var href, text string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" && hasClass(node, "kg-btn") {
			href = ctx.rewriteURL(getAttr(node, "href"))
			text = strings.TrimSpace(textContent(node))
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	if href != "" {
		fmt.Fprintf(ctx.buf, "\n[%s](%s)\n", text, href)
	}
}

func renderProductCard(n *html.Node, ctx *convertContext) {
	var imgSrc, title, href, description string

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if node.Data == "img" && hasClass(node, "kg-product-card-image") {
				imgSrc = ctx.rewriteURL(getAttr(node, "src"))
			}
			if hasClass(node, "kg-product-card-title") {
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && c.Data == "a" {
						href = ctx.rewriteURL(getAttr(c, "href"))
						title = strings.TrimSpace(textContent(c))
						return
					}
				}
				title = strings.TrimSpace(textContent(node))
			}
			if hasClass(node, "kg-product-card-description") {
				description = strings.TrimSpace(textContent(node))
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	ctx.buf.WriteString("\n")
	if imgSrc != "" {
		fmt.Fprintf(ctx.buf, "![%s](%s)\n", title, imgSrc)
	}
	if href != "" {
		fmt.Fprintf(ctx.buf, "**[%s](%s)**\n", title, href)
	} else if title != "" {
		fmt.Fprintf(ctx.buf, "**%s**\n", title)
	}
	if description != "" {
		fmt.Fprintf(ctx.buf, "%s\n", description)
	}
}

func renderFileCard(n *html.Node, ctx *convertContext) {
	var href, title, filename, filesize string

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if node.Data == "a" && hasClass(node, "kg-file-card-container") {
				href = ctx.rewriteURL(getAttr(node, "href"))
			}
			if hasClass(node, "kg-file-card-title") {
				title = strings.TrimSpace(textContent(node))
			}
			if hasClass(node, "kg-file-card-filename") {
				filename = strings.TrimSpace(textContent(node))
			}
			if hasClass(node, "kg-file-card-filesize") {
				filesize = strings.TrimSpace(textContent(node))
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	if href == "" {
		return
	}

	label := title
	if label == "" {
		label = filename
	}
	if filename != "" && filesize != "" {
		label = fmt.Sprintf("%s (%s, %s)", label, filename, filesize)
	}

	fmt.Fprintf(ctx.buf, "\n[%s](%s)\n", label, href)
}
