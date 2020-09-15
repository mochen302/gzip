package gzipx

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"strings"
)

type gzipHandler struct {
	*Options
	CompressLevel int
}

func newGzipHandler(level int, options ...Option) *gzipHandler {
	handler := &gzipHandler{
		Options:       DefaultOptions,
		CompressLevel: level,
	}
	for _, setter := range options {
		setter(handler.Options)
	}
	return handler
}

func (g *gzipHandler) Handle(c *gin.Context) {
	if fn := g.DecompressFn; fn != nil {
		fn(c)
	}

	needCompress := g.shouldCompress(c.Request)
	compressMinLength := int32(2000000000)
	if g.FuncResponseCompressMinLength != nil {
		compressMinLength = g.FuncResponseCompressMinLength()
	}

	gzipWriter := &gzipWriter{
		compressMinLength: compressMinLength,
		ResponseWriter:    c.Writer,
		compressLevel:     g.CompressLevel,
		needCompress:      needCompress,
	}
	c.Writer = gzipWriter

	defer gzipWriter.close()

	if reader, exists := c.Get(DecompressReader); exists {
		gzipReader := reader.(*gzip.Reader)
		defer func() {
			if err := recover(); err != nil {
				println(err)
			}
			err := gzipReader.Close()
			if err != nil {
				println(err)
			}
		}()
	}

	c.Next()
}

func (g *gzipHandler) shouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {

		return false
	}

	extension := filepath.Ext(req.URL.Path)
	if g.ExcludedExtensions.Contains(extension) {
		return false
	}

	if g.ExcludedPaths.Contains(req.URL.Path) {
		return false
	}
	if g.ExcludedPathesRegexs.Contains(req.URL.Path) {
		return false
	}

	return true
}
