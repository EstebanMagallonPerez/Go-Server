package main

import (
	"log"
	"net/http"
	"fmt"
	"strings"
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"time"
	"encoding/json"
)

type folderProperties struct {
	Header string
	Footer string
	Compress bool
}

type httpRequestData struct {
	writer http.ResponseWriter
	request  *http.Request
	fileName string
	fileType string
	folderPath string
	properties folderProperties
}


var port = ":8080"

//define directories for the compressed and uncompressed files
var baseDirectory = "./site"
var compressedDirectory = "./compressed"

//System Defaults
var default_compress = true
var default_header   = "/header.html"
var default_footer   = "/footer.html"


//initiate Response will determine if we should attempt to return a gzipped version of the file, or just send it as it is
func parseUrl(data *httpRequestData){
	var url  = data.request.URL.Path
	if url == "/"{
		url = "/index.html"
	}
	extensionIndex := strings.LastIndex(url, ".")
	fileIndex:= strings.LastIndex(url, "/")
	data.fileType = url[extensionIndex:len(url)]
	data.fileName = url[fileIndex+1:extensionIndex]
	data.folderPath = url[:fileIndex+1]
	/*
		TODO
		Move this logic to a folder level so we do not need to pull the properties every time a file is accessed
	*/
	//Parse JSON, and store results into the requestData
	content, err := ioutil.ReadFile(baseDirectory+data.folderPath + "_folder.json")
	if err != nil {
		data.properties.Header = default_header
		data.properties.Footer = default_footer
		data.properties.Compress = default_compress
		fmt.Printf("%+v\n",data)
		//there is no _folder.json treat everything as default
		return
	}
	json.Unmarshal(content, &data.properties)
}


func initiateResponse(data *httpRequestData){
	fmt.Printf("initiateResponse\n")
	parseUrl(data)
	if data.properties.Compress && strings.Contains(data.request.Header["Accept-Encoding"][0], "gzip"){
		initiateGzip(data)
	}else{
		resolve(data)
	}
}

func initiateGzip(data *httpRequestData){
	fmt.Println("initiateGzip")
	compFile, err := os.Stat(compressedDirectory + data.folderPath + data.fileName + data.fileType + ".gz")
	//if compressed file doesnt exist... make it and serve the uncompressed file for now
	if err != nil {
		go buildGzip(data)
		serveGzip(data)
		return
	}
	uncompFile, err := os.Stat(baseDirectory + data.folderPath + data.fileName + data.fileType)
	if err != nil {
		//return 404
		fmt.Println(err)
		return
	}
	//if compressed is newer than the uncompressed file
	if compFile.ModTime().After(uncompFile.ModTime()){
		serveGzip(data)
	}else{
		go buildGzip(data)
		resolve(data)
	}
}

func serveGzip(data *httpRequestData){
	fmt.Println("serveGzip")
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
	fmt.Println("buildGzip")
	if _, err := os.Stat(compressedDirectory+data.folderPath); os.IsNotExist(err) {
		fmt.Println("making the directory")
		os.MkdirAll(compressedDirectory+data.folderPath, 0444)
	}

	rawbytes, err := ioutil.ReadFile(baseDirectory+data.folderPath + data.fileName+data.fileType)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if data.fileType == ".html"{
		fmt.Println("adding header and footer")
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

//this builds the full file from the template files
func buildHTML(){
	fmt.Printf("buildHTML\n")
}

func resolve(data *httpRequestData){
	rawbytes, err := ioutil.ReadFile(baseDirectory+data.folderPath + data.fileName+data.fileType)
	fmt.Printf("resolve\n")
	if data.fileType == ".html"{
		fmt.Println("adding header and footer")
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

	outputUrl:= baseDirectory+data.folderPath + data.fileName + data.fileType
	fmt.Println(outputUrl)
	http.ServeFile(data.writer, data.request, outputUrl)
}

func main() {
	if _, err := os.Stat(baseDirectory); os.IsNotExist(err) {
		fmt.Println("making the directory")
		os.MkdirAll(baseDirectory, 0444)
	}
	if _, err := os.Stat(compressedDirectory); os.IsNotExist(err) {
		fmt.Println("making the directory")
		os.MkdirAll(compressedDirectory, 0444)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		s := httpRequestData{writer: w, request: req}
		initiateResponse(&s);
	})
	err := http.ListenAndServeTLS(port, "cert.pem", "key.unencrypted.pem", nil)
	log.Fatal(err)
}