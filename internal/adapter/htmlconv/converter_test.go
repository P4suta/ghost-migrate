package htmlconv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConverter_Convert(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			"empty",
			"",
			"",
		},
		{
			"paragraph",
			"<p>Hello world</p>",
			"Hello world\n",
		},
		{
			"headings",
			"<h1>Title</h1><h2>Subtitle</h2>",
			"# Title\n\n## Subtitle\n",
		},
		{
			"bold and italic",
			"<p><strong>bold</strong> and <em>italic</em></p>",
			"**bold** and *italic*\n",
		},
		{
			"link",
			`<p><a href="https://example.com">click</a></p>`,
			"[click](https://example.com)\n",
		},
		{
			"image",
			`<img src="test.jpg" alt="photo">`,
			"![photo](test.jpg)\n",
		},
		{
			"code block",
			`<pre><code class="language-go">func main() {}</code></pre>`,
			"```go\nfunc main() {}\n```\n",
		},
		{
			"unordered list",
			"<ul><li>a</li><li>b</li></ul>",
			"- a\n- b\n",
		},
		{
			"ordered list",
			"<ol><li>first</li><li>second</li></ol>",
			"1. first\n2. second\n",
		},
		{
			"blockquote",
			"<blockquote><p>quoted text</p></blockquote>",
			"> quoted text\n",
		},
		{
			"horizontal rule",
			"<hr>",
			"---\n",
		},
		{
			"inline code",
			"<p>use <code>fmt.Println</code> here</p>",
			"use `fmt.Println` here\n",
		},
		{
			"strikethrough",
			"<p><del>old</del></p>",
			"~~old~~\n",
		},
		{
			"table",
			"<table><tr><th>A</th><th>B</th></tr><tr><td>1</td><td>2</td></tr></table>",
			"| A | B |\n| --- | --- |\n| 1 | 2 |\n",
		},
	}

	c := NewConverter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.Convert(tt.html, nil)
			if err != nil {
				t.Fatalf("Convert() error: %v", err)
			}
			if tt.html == "" {
				if got != "" {
					t.Errorf("Convert(\"\") = %q, want \"\"", got)
				}
				return
			}
			if strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("Convert()\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestGhostCards(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			"image card",
			`<figure class="kg-card kg-image-card"><img src="photo.jpg" alt="desc"><figcaption>Caption</figcaption></figure>`,
			"![desc](photo.jpg)\n*Caption*\n",
		},
		{
			"image card without caption",
			`<figure class="kg-card kg-image-card"><img src="photo.jpg" alt="desc"></figure>`,
			"![desc](photo.jpg)\n",
		},
		{
			"bookmark card",
			`<figure class="kg-card kg-bookmark-card"><a class="kg-bookmark-container" href="https://example.com"><div class="kg-bookmark-content"><div class="kg-bookmark-title">Title</div><div class="kg-bookmark-description">Desc</div></div></a></figure>`,
			"> **[Title](https://example.com)**\n> Desc\n",
		},
		{
			"callout card",
			`<div class="kg-card kg-callout-card"><div class="kg-callout-emoji">💡</div><div class="kg-callout-text">Info</div></div>`,
			"> 💡 Info\n",
		},
		{
			"toggle card",
			`<div class="kg-card kg-toggle-card"><div class="kg-toggle-heading-text"><h4>Q</h4></div><div class="kg-toggle-content"><p>A</p></div></div>`,
			"<details>\n<summary>Q</summary>\n\nA\n\n</details>\n",
		},
		{
			"button card",
			`<div class="kg-card kg-button-card"><a href="https://example.com" class="kg-btn">Click</a></div>`,
			"[Click](https://example.com)\n",
		},
		{
			"embed card",
			`<figure class="kg-card kg-embed-card"><iframe src="https://youtube.com/embed/xxx"></iframe><figcaption>Video</figcaption></figure>`,
			"https://youtube.com/embed/xxx\n*Video*\n",
		},
		{
			"gallery card",
			`<figure class="kg-card kg-gallery-card"><div class="kg-gallery-container"><div class="kg-gallery-row"><div class="kg-gallery-image"><img src="a.jpg" alt="A"></div><div class="kg-gallery-image"><img src="b.jpg" alt="B"></div></div></div></figure>`,
			"![A](a.jpg)\n\n![B](b.jpg)\n",
		},
		{
			"product card",
			`<div class="kg-card kg-product-card"><div class="kg-product-card-container"><img src="product.jpg" class="kg-product-card-image"><div class="kg-product-card-title-container"><h4 class="kg-product-card-title"><a href="https://shop.com/item">Product</a></h4></div><div class="kg-product-card-description"><p>Great product</p></div></div></div>`,
			"![Product](product.jpg)\n**[Product](https://shop.com/item)**\nGreat product\n",
		},
		{
			"file card",
			`<div class="kg-card kg-file-card"><a class="kg-file-card-container" href="https://example.com/file.pdf"><div class="kg-file-card-contents"><div class="kg-file-card-title">My File</div><div class="kg-file-card-metadata"><div class="kg-file-card-filename">file.pdf</div><div class="kg-file-card-filesize">1.2 MB</div></div></div></a></div>`,
			"[My File (file.pdf, 1.2 MB)](https://example.com/file.pdf)\n",
		},
	}

	c := NewConverter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.Convert(tt.html, nil)
			if err != nil {
				t.Fatalf("Convert() error: %v", err)
			}
			if strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("Convert()\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestConverter_URLRewriter(t *testing.T) {
	rewriter := func(url string) string {
		if strings.Contains(url, "__GHOST_URL__/content/images/") {
			return strings.Replace(url, "__GHOST_URL__/content/images/2024/photo.jpg", "photo.jpg", 1)
		}
		return url
	}

	tests := []struct {
		name string
		html string
		want string
	}{
		{
			"img src rewritten",
			`<img src="__GHOST_URL__/content/images/2024/photo.jpg" alt="pic">`,
			"![pic](photo.jpg)\n",
		},
		{
			"link href rewritten",
			`<p><a href="__GHOST_URL__/content/images/2024/photo.jpg">link</a></p>`,
			"[link](photo.jpg)\n",
		},
		{
			"image card rewritten",
			`<figure class="kg-card kg-image-card"><img src="__GHOST_URL__/content/images/2024/photo.jpg" alt="pic"></figure>`,
			"![pic](photo.jpg)\n",
		},
		{
			"non-matching URL unchanged",
			`<p><a href="https://example.com">link</a></p>`,
			"[link](https://example.com)\n",
		},
	}

	c := NewConverter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.Convert(tt.html, rewriter)
			if err != nil {
				t.Fatalf("Convert() error: %v", err)
			}
			if strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("Convert()\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestGoldenFiles(t *testing.T) {
	goldenDir := "../../../testdata/golden"
	entries, err := os.ReadDir(goldenDir)
	if err != nil {
		t.Skip("no golden directory")
	}

	c := NewConverter()
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".html")
		t.Run(name, func(t *testing.T) {
			htmlBytes, err := os.ReadFile(filepath.Join(goldenDir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			got, err := c.Convert(string(htmlBytes), nil)
			if err != nil {
				t.Fatal(err)
			}

			mdPath := filepath.Join(goldenDir, name+".md")
			if os.Getenv("UPDATE_GOLDEN") != "" {
				os.WriteFile(mdPath, []byte(got), 0o644)
				return
			}

			wantBytes, err := os.ReadFile(mdPath)
			if err != nil {
				t.Fatalf("no golden file %s (set UPDATE_GOLDEN=1 to create)", mdPath)
			}
			if got != string(wantBytes) {
				t.Errorf("output differs from golden file\ngot:\n%s\nwant:\n%s", got, string(wantBytes))
			}
		})
	}
}
