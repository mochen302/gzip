package gzip

import (
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
	gzipWriter := &gzipWriter{
		compressMinLength: g.ResponseCompressMinLength,
		ResponseWriter:    c.Writer,
		compressLevel:     g.CompressLevel,
		needCompress:      needCompress,
	}
	c.Writer = gzipWriter

	defer gzipWriter.flushThenClose()

	c.Next()
}

func (g *gzipHandler) shouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") ||
		strings.Contains(req.Header.Get("Content-Type"), "text/event-stream") {

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
