package main

import(
	"net/http"
	"fmt"
	"io/ioutil"
	"os"
	"time"
	"bytes"
	"compress/gzip"
)
func serveGzip(data *httpRequestData){
	data.writer.Header().Set("Content-Type", data.contentType)
	data.writer.Header().Set("Content-Encoding", "gzip")
	compFile, err := os.Open(compressedDirectory+data.folderPath + data.fileName + data.fileType+".gz")
	if err != nil {
		fmt.Println(err)
	}
	defer compFile.Close()
	http.ServeContent(data.writer, data.request, data.folderPath + data.fileName + data.fileType, time.Time{}, compFile)
}

//this will build the gzip or just point to the file if it already exists
func buildGzip(data *httpRequestData){
	makeDir(compressedDirectory+data.folderPath)
	rawbytes, err := ioutil.ReadFile(baseDirectory+data.folderPath + data.fileName+data.fileType)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if data.fileType == ".html"{
		//prepend Header
		if data.properties.Header != ""{
			temp, err := ioutil.ReadFile(baseDirectory+data.properties.Header)

			if err != nil {
				fmt.Println(err)
			}else{
				rawbytes = append(temp,rawbytes...)
			}

		}
		if data.properties.Footer != ""{
			temp, err := ioutil.ReadFile(baseDirectory+data.properties.Footer)

			if err != nil {
				fmt.Println(err)
			}else{
				rawbytes = append(rawbytes,temp...)
			}
		}
	}

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	writer.Write(rawbytes)
	writer.Close()

	err = ioutil.WriteFile(compressedDirectory+data.folderPath + data.fileName + data.fileType+".gz", buf.Bytes(), 0444)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
