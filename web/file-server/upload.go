package file_server

import (
	"awesomeProject2/web/context"
	"awesomeProject2/web/handler"
	"io"
	"log"
	"os"
	"path"
)

type FileUploader struct {
	dst   string
	field string
}

func (f *FileUploader) Handle() handler.HandleFunc {
	return func(c *context.Context) {
		src, srcHeader, err := c.R.FormFile(f.field)
		defer func() {
			_ = src.Close()
		}()
		if err != nil {
			er := c.SystemErrorJson(err)
			log.Println(er)
			return
		}
		p := path.Join(f.dst, srcHeader.Filename)
		var dst *os.File
		dst, err = os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		defer func() {
			_ = dst.Close()
		}()
		if err != nil {
			er := c.SystemErrorJson(err)
			log.Println(er)
			return
		}
		_, err = io.CopyBuffer(dst, src, nil)
		if err != nil {
			er := c.SystemErrorJson(err)
			log.Println(er)
			return
		}
		_ = c.OkJson("上传成功")
	}
}

func NewFileUploader(dst string, field string) *FileUploader {
	return &FileUploader{
		dst:   dst,
		field: field,
	}
}
