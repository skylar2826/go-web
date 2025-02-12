package file_server

import (
	"awesomeProject2/web/context"
	"awesomeProject2/web/handler"
	"net/http"
	"path"
)

type FileDownloader struct {
	dst   string
	field string
}

func (f *FileDownloader) Handle() handler.HandleFunc {
	return func(c *context.Context) {
		query := c.R.URL.Query()
		fileName := query.Get(f.field)
		p := path.Join(f.dst, fileName)

		header := c.W.Header()
		header.Set("Content-Disposition", "attachment;filename="+fileName)
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")
		header.Set("Cache-Control", "must-revalidate")
		header.Set("Pragma", "public")

		http.ServeFile(c.W, c.R, p)
	}
}

func NewFileDownloader(dst string, field string) *FileDownloader {
	return &FileDownloader{
		dst:   dst,
		field: field,
	}
}
