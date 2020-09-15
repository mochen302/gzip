package gzipx

import (
	"compress/gzip"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	DecompressHeader      = "X-Puzzle-Compress"
	DecompressHeaderValue = "gzip"
	DecompressReader      = "X-Puzzle-Compress-Reader"
)

var (
	DefaultExcludedExtentions = NewExcludedExtensions([]string{
		".png", ".gif", ".jpeg", ".jpg",
	})
	DefaultOptions = &Options{
		ExcludedExtensions: DefaultExcludedExtentions,
	}
)

type Options struct {
	ExcludedExtensions            ExcludedExtensions
	ExcludedPaths                 ExcludedPaths
	ExcludedPathesRegexs          ExcludedPathesRegexs
	FuncResponseCompressMinLength func() int32
	DecompressFn                  func(c *gin.Context)
}

type Option func(*Options)

func WithExcludedExtensions(args []string) Option {
	return func(o *Options) {
		o.ExcludedExtensions = NewExcludedExtensions(args)
	}
}

func WithExcludedPaths(args []string) Option {
	return func(o *Options) {
		o.ExcludedPaths = NewExcludedPaths(args)
	}
}

func WithExcludedPathsRegexs(args []string) Option {
	return func(o *Options) {
		o.ExcludedPathesRegexs = NewExcludedPathesRegexs(args)
	}
}

func WithDecompressFn(decompressFn func(c *gin.Context)) Option {
	return func(o *Options) {
		o.DecompressFn = decompressFn
	}
}

// Using map for better lookup performance
type ExcludedExtensions map[string]bool

func NewExcludedExtensions(extensions []string) ExcludedExtensions {
	res := make(ExcludedExtensions)
	for _, e := range extensions {
		res[e] = true
	}
	return res
}

func (e ExcludedExtensions) Contains(target string) bool {
	_, ok := e[target]
	return ok
}

type ExcludedPaths []string

func NewExcludedPaths(paths []string) ExcludedPaths {
	return ExcludedPaths(paths)
}

func (e ExcludedPaths) Contains(requestURI string) bool {
	for _, path := range e {
		if strings.HasPrefix(requestURI, path) {
			return true
		}
	}
	return false
}

type ExcludedPathesRegexs []*regexp.Regexp

func NewExcludedPathesRegexs(regexs []string) ExcludedPathesRegexs {
	result := make([]*regexp.Regexp, len(regexs), len(regexs))
	for i, reg := range regexs {
		result[i] = regexp.MustCompile(reg)
	}
	return result
}

func (e ExcludedPathesRegexs) Contains(requestURI string) bool {
	for _, reg := range e {
		if reg.MatchString(requestURI) {
			return true
		}
	}
	return false
}

func DefaultDecompressHandle(c *gin.Context) {
	DefaultDecompressHandleWithHeader(c, DecompressHeader, DecompressHeaderValue)
}

func DefaultDecompressHandleWithHeader(c *gin.Context, header, expectHeaderValue string) {
	if c.Request.Body == nil {
		return
	}

	headerValue := c.Request.Header.Get(header)
	if headerValue != expectHeaderValue {
		return
	}

	r, err := gzip.NewReader(c.Request.Body)
	if err != nil {
		println(err)
		return
	}

	c.Set(DecompressReader, r)

	c.Request.Header.Del("Content-Encoding")
	c.Request.Header.Del("Content-Length")
	c.Request.Body = r
}

//func GetReader() *gzip.Reader {
//	buf := &bytes.Buffer{}
//	gz, _ := gzip.NewWriterLevel(buf, gzip.DefaultCompression)
//	if _, err := gz.Write([]byte("this is test")); err != nil {
//		gz.Close()
//	}
//	defer gz.Close()
//
//	reader, err := gzip.NewReader(buf)
//	if err != nil {
//		panic(err)
//	}
//
//	return reader
//}
