package handler

import (
	"awesomeProject2/web/context"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type fileCacheItem struct {
	name        string
	data        []byte
	size        int
	contentType string
}

type StaticResourceHandlerOption func(*StaticResourceHandler)

type StaticResourceHandler struct {
	dir         string
	pathPrefix  string
	extMap      map[string]string
	c           *lru.Cache
	maxFileSize int
}

func (s *StaticResourceHandler) ServeStaticResource(c *context.Context) {
	path := strings.TrimPrefix(c.R.URL.Path, s.pathPrefix)
	path = filepath.Join(s.dir, path)

	pathSlice := strings.Split(path, "/")
	fileName := pathSlice[len(pathSlice)-1]
	file, found := s.getCacheFile(fileName)
	if !found {
		f, err := os.Open(path)
		if err != nil {
			er := c.SystemErrorJson(err)
			if er != nil {
				log.Println(er)
			}
			return
		}

		ext := getExt(f.Name())
		contentType, ok := s.extMap[ext]
		if !ok {
			er := c.SystemErrorJson(errors.New("文件类型不支持传送"))
			if er != nil {
				log.Println(er)
			}
			return
		}

		var data []byte
		data, err = io.ReadAll(f)
		file = &fileCacheItem{
			name:        fileName,
			size:        len(data),
			data:        data,
			contentType: contentType,
		}
		err = s.cacheFile(file)
		if err != nil {
			log.Println(err)
		}
	}

	s.writeFileResponse(c, file)
}

func (s *StaticResourceHandler) writeFileResponse(c *context.Context, file *fileCacheItem) {
	c.W.Header().Set("content-type", file.contentType)
	c.W.Header().Set("content-length", strconv.Itoa(file.size))
	c.OkJsonDirect(file.data)
}

func (s *StaticResourceHandler) cacheFile(file *fileCacheItem) error {
	if s.c != nil && file.size <= s.maxFileSize {
		s.c.Add(file.name, file)
	}
	return errors.New(fmt.Sprintf("缓存文件失败：%s", file.name))
}

func (s *StaticResourceHandler) getCacheFile(fileName string) (*fileCacheItem, bool) {
	if s.c != nil {
		file, ok := s.c.Get(fileName)
		if !ok {
			return nil, false
		}
		return file.(*fileCacheItem), ok
	}

	return nil, false
}

func getExt(fileName string) string {
	arr := strings.Split(fileName, ".")
	return arr[len(arr)-1]
}

func WithMoreStaticResourceExt(extMap map[string]string) StaticResourceHandlerOption {
	return func(s *StaticResourceHandler) {
		if s.extMap == nil {
			s.extMap = make(map[string]string, len(extMap))
		}
		for key, value := range extMap {
			s.extMap[key] = value
		}
	}
}

func WithFileCache(maxFileSize, maxFileCnt int) StaticResourceHandlerOption {
	return func(s *StaticResourceHandler) {
		c, err := lru.New(maxFileCnt)
		if err != nil {
			log.Println(err)
			return
		}
		s.c = c
		s.maxFileSize = maxFileSize
	}
}

func NewStaticResourceHandler(dir string, pathPrefix string, options ...StaticResourceHandlerOption) *StaticResourceHandler {
	srh := &StaticResourceHandler{
		dir:        dir,
		pathPrefix: pathPrefix,
		extMap: map[string]string{
			"png":  "image/png",
			"jpg":  "image/jpg",
			"jpeg": "image/jpeg",
		},
	}

	for _, option := range options {
		option(srh)
	}

	return srh
}
