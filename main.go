package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/dustin/go-humanize"
	"path/filepath"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"bufio"
	"io"
)

func main() {
	router := gin.Default()
	// set a limit for uploads
	// router.MaxMultipartMemory = 100 << 20
	
	router.Use(cors.New(cors.Config{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{"POST", "PUT", "PATCH", "DELETE"},
        AllowHeaders: []string{"Content-Type,access-control-allow-origin, access-control-allow-headers"},
    }))
	
	router.POST("/upload", func(c *gin.Context) {
		if(c.Query("dir") != ""){
			var dir = c.Query("dir")
			file, header, err := c.Request.FormFile("file")
			if err != nil {
				c.String(http.StatusBadRequest, fmt.Sprintf("file err : %s", err.Error()))
				return
			}
		
			fileExt := filepath.Ext(header.Filename)
			originalFileName := strings.TrimSuffix(filepath.Base(header.Filename), filepath.Ext(header.Filename))
			filename := originalFileName + fileExt
		
			out, _ := os.Create(dir +"/"+ filename)
			defer out.Close()
			io.Copy(out, file)
			c.String(http.StatusOK,"uploaded on " + dir +"/"+ filename + " you can now return")
		}else{
			c.String(http.StatusOK,"spesify location")
		}
	})

	router.GET("/file/*path", func(c *gin.Context) {
		byteFile, _ := os.Open(c.Param("path"))
		defer byteFile.Close()
		stats, _ := byteFile.Stat()

	
		var size int64 = stats.Size()
		bytes := make([]byte, size)
	
		bufr := bufio.NewReader(byteFile)
		bufr.Read(bytes)
		mimeType := http.DetectContentType(bytes)
		c.Data(http.StatusOK, mimeType, bytes)
	})

	router.GET("/", func(c *gin.Context) {
		var dir,_ = os.Getwd()

		if(c.Query("f") != ""){
			dir = c.Query("f")
		}

		files, _ := ioutil.ReadDir(dir)

		var dirArr = strings.Split(dir, "/");
		var upDir = strings.Join(dirArr[:len(dirArr) - 1], "/");
		if(upDir == ""){
			upDir = "/"
		}
		var list = fmt.Sprintf("<a href='/?f=%s'>../</a>", upDir);
		
		for _, f := range files {
			if(f.IsDir()){
				list = list + fmt.Sprintf("<li><a href='/?f=%s'>%s 📁</a></li>", strings.Join(append(dirArr, f.Name()), "/"), f.Name())
			}else{
				list = list + fmt.Sprintf("<li><a href='/file/%v'>%s 📄 (Size: %v)</a></li>",strings.Join(append(dirArr, f.Name()), "/") , f.Name(), humanize.Bytes(uint64(f.Size())), )
			}
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK,fmt.Sprintf(`<ol>%s</ol>
		<form action="/upload?dir=`+strings.Join(dirArr, "/")+`" enctype="multipart/form-data" method="post">
			<input type="file" name="file">
			<input type="submit">
	  	</form>`, list))
	})

	router.Run(":8080")
}