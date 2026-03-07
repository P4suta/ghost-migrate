package port

type URLRewriter func(string) string

type ContentConverter interface {
	Convert(html string, rewriter URLRewriter) (string, error)
}
