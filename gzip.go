package gzip

import (
	"compress/gzip"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

const (
	BestCompression    = gzip.BestCompression
	BestSpeed          = gzip.BestSpeed
	DefaultCompression = gzip.DefaultCompression
	NoCompression      = gzip.NoCompression
)

func Gzip(level int, options ...Option) gin.HandlerFunc {
	return newGzipHandler(level, options...).Handle
}

type gzipWriter struct {
	compressMinLength int32
	forceCompress     bool

	alreadyWriteLength int32
	isCompress         bool

	compressLevel int

	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	if g.tryCompress(int32(len(s))) {
		bytes, err := g.writer.Write([]byte(s))
		if err == nil {
			defer func() {
				if err := recover(); err != nil {
					//ignore
				}
				_ = g.writer.Close()
			}()
			g.ResponseWriter.WriteHeaderNow()
		}
		return bytes, err
	} else {
		return g.ResponseWriter.Write([]byte(s))
	}
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	if g.tryCompress(int32(len(data))) {
		bytes, err := g.writer.Write(data)
		if err == nil {
			defer func() {
				if err := recover(); err != nil {
					//ignore
				}
				_ = g.writer.Close()
			}()
			g.ResponseWriter.WriteHeaderNow()
		}
		return bytes, err
	} else {
		return g.ResponseWriter.Write(data)
	}
}

func (g *gzipWriter) tryCompress(currentLength int32) bool {
	g.alreadyWriteLength = currentLength

	if g.alreadyWriteLength >= g.compressMinLength && g.forceCompress {
		g.isCompress = true

		writer, err := gzip.NewWriterLevel(ioutil.Discard, g.compressLevel)
		if err != nil {
			g.isCompress = false
		} else {
			g.tryWriteHeaders()
			g.writer = writer
		}
	}

	return g.isCompress
}

// Fix: https://github.com/mholt/caddy/issues/38
func (g *gzipWriter) WriteHeader(code int) {
	g.tryWriteHeaders()
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipWriter) tryWriteHeaders() {
	if g.isCompress {
		header := g.ResponseWriter.Header()
		header.Set("Content-Encoding", "gzip")
		header.Set("Vary", "Accept-Encoding")
		header.Set("Content-Length", fmt.Sprint(g.alreadyWriteLength))
	}
}
