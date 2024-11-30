package main

import (
	"flag"
	"fmt"
	"os"
	"text/template"
)

type Params struct {
	Version string
}

func main() {
	var params Params
	flag.StringVar(&params.Version, "version", "latest", "version to install")

	var templateFile string
	flag.StringVar(&templateFile, "path", "", "path to the template file")

	var path string
	flag.StringVar(&path, "out", "", "output path")

	flag.Parse()

	if path == "" {
		path = "."
	}

	if templateFile == "" {
		panic("template is required")
	}

	if params.Version == "" {
		panic("version is required")
	}

	fmt.Printf("Compiling installer script(%q) with version: %q\n", templateFile, params.Version)

	template, err := template.ParseFiles(templateFile)
	if err != nil {
		panic(err)
	}

	installer := fmt.Sprintf("%s/clio-installer.sh", path)
	fo, err := os.Create(installer)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	err = template.Execute(fo, params)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Installer compiled %q\n", installer)
}
