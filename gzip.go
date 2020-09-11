package gzip

import (
	"compress/gzip"
	"fmt"
	"github.com/gin-gonic/gin"
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
	needCompress      bool

	alreadyWriteLength int32
	isCompress         bool

	compressLevel int

	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	byes0, err := g.writeByes0([]byte(s))
	if err == nil {
		defer g.flush()
	}
	return byes0, err
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	byes0, err := g.writeByes0(data)
	if err == nil {
		defer g.flush()
	}
	return byes0, err
}

func (g *gzipWriter) writeByes0(bytes []byte) (int, error) {
	if g.tryCompress(int32(len(bytes))) {
		g.tryWriteHeaders()

		byteLength, err := g.writer.Write(bytes)
		if err == nil {
			g.alreadyWriteLength = int32(byteLength)
		}
		return byteLength, err
	} else {
		return g.ResponseWriter.Write(bytes)
	}
}

func (g *gzipWriter) flush() {
	if !g.isCompress {
		return
	}
	err := g.writer.Flush()
	if err != nil {
		fmt.Println(err)
	}
}

func (g *gzipWriter) close() {
	if !g.isCompress {
		return
	}

	func() {
		if err := recover(); err != nil {
			//ignore
		}
		_ = g.writer.Close()
		g.tryWriteHeaderWriteLength()
	}()
}

func (g *gzipWriter) tryCompress(currentLength int32) bool {
	g.alreadyWriteLength = currentLength

	if g.alreadyWriteLength >= g.compressMinLength && g.needCompress {
		g.isCompress = true

		writer, err := gzip.NewWriterLevel(g.ResponseWriter, g.compressLevel)
		if err != nil {
			g.isCompress = false
		} else {
			g.writer = writer
		}
	}

	return g.isCompress
}

// Fix: https://github.com/mholt/caddy/issues/38
func (g *gzipWriter) WriteHeader(code int) {
	g.ResponseWriter.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipWriter) tryWriteHeaders() {
	if g.isCompress {
		header := g.ResponseWriter.Header()
		header.Set("Content-Encoding", "gzip")
		header.Set("Vary", "Accept-Encoding")
	}
}

func (g *gzipWriter) tryWriteHeaderWriteLength() {
	if g.isCompress {
		header := g.ResponseWriter.Header()
		header.Set("Content-Length", fmt.Sprint(g.alreadyWriteLength))
	}
}
