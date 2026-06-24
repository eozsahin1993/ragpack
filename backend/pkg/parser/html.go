package parser

import (
	"context"
	"io"
	"iter"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var multipleBlankLines = regexp.MustCompile(`\n{3,}`)

// HTMLParser converts HTML to Markdown and streams sections via MarkdownParser.
type HTMLParser struct{}

// skipTags are subtrees that contain no useful content for RAG.
var skipTags = map[string]bool{
	"script": true, "style": true, "nav": true, "footer": true,
	"aside": true, "header": true, "form": true, "button": true, "iframe": true,
}

func (p *HTMLParser) Parse(ctx context.Context, r io.ReadCloser) iter.Seq2[Unit, error] {
	return func(yield func(Unit, error) bool) {
		defer r.Close()

		doc, err := html.Parse(r)
		if err != nil {
			yield(Unit{}, err)
			return
		}

		var sb strings.Builder
		htmlToMarkdown(doc, &sb, 0)
		md := strings.TrimSpace(multipleBlankLines.ReplaceAllString(sb.String(), "\n\n"))
		if md == "" {
			return
		}

		mp := &MarkdownParser{}
		for unit, err := range mp.Parse(ctx, io.NopCloser(strings.NewReader(md))) {
			if !yield(unit, err) {
				return
			}
		}
	}
}

func htmlToMarkdown(n *html.Node, sb *strings.Builder, listDepth int) {
	if n.Type == html.ElementNode && skipTags[n.Data] {
		return
	}
	if n.Type == html.TextNode {
		if text := strings.TrimSpace(n.Data); text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
		return
	}
	if n.Type != html.ElementNode {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlToMarkdown(c, sb, listDepth)
		}
		return
	}

	tag := n.Data
	switch tag {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := int(tag[1] - '0')
		sb.WriteString("\n\n")
		sb.WriteString(strings.Repeat("#", level))
		sb.WriteString(" ")
		var inner strings.Builder
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlToMarkdown(c, &inner, listDepth)
		}
		sb.WriteString(strings.TrimSpace(inner.String()))
		sb.WriteString("\n\n")

	case "p", "div", "section", "article", "blockquote":
		sb.WriteString("\n\n")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlToMarkdown(c, sb, listDepth)
		}
		sb.WriteString("\n\n")

	case "br":
		sb.WriteString("\n")

	case "ul", "ol":
		sb.WriteString("\n")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlToMarkdown(c, sb, listDepth+1)
		}
		sb.WriteString("\n")

	case "li":
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat("  ", listDepth-1))
		sb.WriteString("- ")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlToMarkdown(c, sb, listDepth)
		}

	case "pre":
		sb.WriteString("\n\n```\n")
		var inner strings.Builder
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlCollectText(c, &inner)
		}
		sb.WriteString(strings.TrimSpace(inner.String()))
		sb.WriteString("\n```\n\n")

	case "code":
		var inner strings.Builder
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlCollectText(c, &inner)
		}
		if text := strings.TrimSpace(inner.String()); text != "" {
			sb.WriteString("`")
			sb.WriteString(text)
			sb.WriteString("`")
		}

	case "strong", "b":
		sb.WriteString("**")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlToMarkdown(c, sb, listDepth)
		}
		sb.WriteString("**")

	case "em", "i":
		sb.WriteString("*")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlToMarkdown(c, sb, listDepth)
		}
		sb.WriteString("*")

	case "a":
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlToMarkdown(c, sb, listDepth)
		}

	case "img":
		for _, a := range n.Attr {
			if a.Key == "alt" && strings.TrimSpace(a.Val) != "" {
				sb.WriteString(strings.TrimSpace(a.Val))
				sb.WriteString(" ")
			}
		}

	case "table":
		sb.WriteString("\n\n")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlToMarkdown(c, sb, listDepth)
		}
		sb.WriteString("\n\n")

	case "tr":
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlToMarkdown(c, sb, listDepth)
		}
		sb.WriteString("\n")

	case "td", "th":
		sb.WriteString(" ")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlToMarkdown(c, sb, listDepth)
		}
		sb.WriteString(" |")

	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlToMarkdown(c, sb, listDepth)
		}
	}
}

func htmlCollectText(n *html.Node, sb *strings.Builder) {
	if n.Type == html.TextNode {
		sb.WriteString(n.Data)
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		htmlCollectText(c, sb)
	}
}
