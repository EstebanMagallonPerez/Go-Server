package main

import (
	"log"
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
	"os"
	"time"
	"encoding/json"
	"mime"
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
	contentType string
	properties folderProperties
}


var port = ":443"

//define directories for the compressed and uncompressed files
var baseDirectory = "./site"
var compressedDirectory = "./compressed"

//System Defaults
var default_compress = true
var default_header   = "/header.html"
var default_footer   = "/footer.html"


var gzipBuff circularBuffer
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
	data.contentType = mime.TypeByExtension(data.fileType)

	/*
		TODO
		Move this logic to a folder level so we do not need to pull the properties every time a file is accessed
	*/
	//Parse JSON, and store results into the requestData
	content, err := ioutil.ReadFile(baseDirectory+data.folderPath + "_folder.json")
	if err != nil {
		//there is no _folder.json treat everything as default
		data.properties.Header = default_header
		data.properties.Footer = default_footer
		data.properties.Compress = default_compress
		return
	}
	json.Unmarshal(content, &data.properties)
}


func initiateResponse(data *httpRequestData){
	parseUrl(data)
	//if file does not exist... dont bother with anything else, and return the 404 page
	uncompFile, err := os.Stat(baseDirectory + data.folderPath + data.fileName + data.fileType)
	if err != nil {
		//return 404
		if data.fileType == ".html"{
			data.writer.WriteHeader(http.StatusNotFound)
			http.ServeFile(data.writer, data.request, baseDirectory+"/404.html")
		}else{
			http.NotFound(data.writer,data.request)
		}
		return
	}
	if data.properties.Compress && strings.Contains(data.request.Header["Accept-Encoding"][0], "gzip"){
		initiateGzip(data,uncompFile.ModTime())
	}else{
		resolve(data)
	}
}

func initiateGzip(data *httpRequestData, uncompTime time.Time){
	compFile, err := os.Stat(compressedDirectory + data.folderPath + data.fileName + data.fileType + ".gz")
	//if compressed file doesnt exist... make it and serve the uncompressed file for now
	if err != nil || compFile.ModTime().Before(uncompTime){
		gzipBuff.insert(data)
		resolve(data)
		return
	}else{
		serveGzip(data)
	}
}



//this builds the full file from the template files
func buildHTML(){
	fmt.Printf("buildHTML\n")
}

func resolve(data *httpRequestData){
	rawbytes, err := ioutil.ReadFile(baseDirectory+data.folderPath + data.fileName+data.fileType)
	if err != nil {
		fmt.Println(err)
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
	data.writer.Header().Set("Content-Type", data.contentType)
	data.writer.Write(rawbytes)
}

func makeDir(path string){
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0444)
	}
}

func gzipHandler(){
	//This runs on a separate thread, and will gzip files when the main thread requests it
	//This handles scenarios where the same file is requested for compression in multiple scenarios, and will force only 1 compression request to be processed for that file instead of multiple.
	for true{
		if !gzipBuff.empty(){
			temp :=gzipBuff.get()
			fmt.Println(temp.fileName)
			fmt.Println("starting the compression")
			buildGzip(temp)
			fmt.Println("finished the compression")
			gzipBuff.pop()
		}
	}
}
func main() {
	makeDir(baseDirectory)
	makeDir(compressedDirectory)
	gzipBuff.init(25)
	go gzipHandler()
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		s := httpRequestData{writer: w, request: req}
		initiateResponse(&s);
	})
	err := http.ListenAndServeTLS(port, "cert.pem", "key.unencrypted.pem", nil)
	log.Fatal(err)
}
