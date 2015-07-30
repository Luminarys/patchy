package main

import (
	"archive/zip"
	"fmt"
	"github.com/hoisie/web"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

//Handles POST upload requests. the updateURL is used to pass messages
//to the urlHandler indicating that the DB should be updated.
func handleUpload(ctx *web.Context, l *library) string {
	//TODO: Implemente limits with settings.ini or something
	err := ctx.Request.ParseMultipartForm(150 * 1024 * 1024)
	if err != nil {
		return "Error handling form!"
	}
	form := ctx.Request.MultipartForm

	//Loop through and append to the response struct
	ul := form.File["file"][0]
	fmt.Println(form.File)
	name := ul.Filename
	fmt.Println(name)
	file, err := ul.Open()
	size, _ := file.Seek(0, 2)
	if size > 150*1024*1024 {
		return "File too big!"
	}
	file.Seek(0, 0)

	f, err := os.Create("temp/" + name)
	if err != nil {
		return "Could not store file!"
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return "Error reading the file"
	}

	//Get MIME
	reader, err := ioutil.ReadAll(f)
	if err != nil {
		return "Error reading file!"
	}
	mime := http.DetectContentType(reader)
	if mime != "application/zip" {
		fmt.Println(mime)
		//return "Make sure this is a ZIP file!"
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		return "Error reading the file"
	}

	_, err = io.Copy(f, file)
	if err != nil {
		return "Error saving the file"
	}
	f.Close()
	defer os.Remove("temp/" + name)

	r, err := zip.OpenReader("temp/" + name)
	if err != nil {
		os.Remove("temp/" + name)
		return "Error opening the file, make sure it is a zip!"
	}
	defer r.Close()

	dir := r.Reader.File[0]

	name = dir.Name
	if !dir.FileInfo().IsDir() {
		return "The zip should contain a single directory!"
	}
	os.Mkdir(musicDir+"/"+name, dir.Mode())

	for _, song := range r.Reader.File[1:] {
		zipped, err := song.Open()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer zipped.Close()

		// get the individual file name and extract the current directory
		path := filepath.Join(musicDir, "/", song.Name)
		if song.FileInfo().IsDir() {
			os.MkdirAll(path, song.Mode())
			fmt.Println("Creating directory", path)
		} else {
			writer, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, song.Mode())

			if err != nil {
				os.Remove(musicDir + "/" + name)
				return "Couldn't unzip file!"
			}

			defer writer.Close()

			if _, err = io.Copy(writer, zipped); err != nil {
				os.Remove(musicDir + "/" + name)
				return "Couldn't unzip file!"
			}
		}
	}
	if err = l.update(); err != nil {
		fmt.Println(err)
		os.Remove(musicDir + "/" + name)
		return "Couldn't update library"
	}
	return "Added!"
}
