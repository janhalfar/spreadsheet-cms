package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

type templateData []map[string]string

func must(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
		os.Exit(1)
	}
}

func readData(source io.Reader) (td templateData) {
	r := csv.NewReader(source)
	i := 0
	td = templateData{}
	header := []string{}
ReadLoop:
	for {
		line, errRead := r.Read()
		if errRead != nil {
			if errRead == io.EOF {
				fmt.Println("done reading csv")
				break ReadLoop
			}
		}
		for trimI, linteItem := range line {
			line[trimI] = strings.Trim(linteItem, " 	")
		}
		i++
		if i == 1 {
			// header line
			header = line
			continue
		}
		templateDataRow := map[string]string{}
		for ii, lineItem := range line {
			if ii < len(header) {
				templateDataRow[header[ii]] = lineItem
			} else {
				must("invalid csv", errors.New("no header for col: "+fmt.Sprint("col", ii, "in row", i)))
			}
		}
		td = append(td, templateDataRow)
	}
	return td
}

func getTemplateFuncs(assetDir string) template.FuncMap {
	return template.FuncMap{
		"List": func(lines string) []string {
			cleanLines := []string{}
			for _, dirtyLine := range strings.Split(lines, "\n") {
				cleanLine := strings.Trim(dirtyLine, " 	")
				if cleanLine == "" {
					continue
				}
				cleanLines = append(cleanLines, cleanLine)
			}
			return cleanLines
		},
		"HasAsset": func(asset string) bool {
			if assetDir == "" {
				flag.Usage()
				must("arg missing", errors.New("if you want to work with assets please define asset-dir"))
			}
			imgName := filepath.Join(assetDir, asset)
			info, errStat := os.Stat(imgName)
			return errStat == nil && !info.IsDir()
		},
		"Empty": func(value interface{}) bool {
			switch value.(type) {
			case string:
				stringValue := value.(string)
				return "" == strings.Trim(stringValue, "	 ")
			default:
				return false
			}
		},
	}
}

func renderData(td templateData, languages []string, outDir string, tpl *template.Template) {
	for i, templateDataRow := range td {
		id, okID := templateDataRow["id"]
		if !okID {
			must("no id", errors.New("no id given"))
		}
		for _, lang := range languages {
			title := id + "-" + lang + ".html"
			suffix := "-" + lang
			languageData := map[string]interface{}{"id": id, "language": lang, "languages": languages}
		FieldLoop:
			for fieldName, data := range templateDataRow {
				if strings.HasSuffix(fieldName, suffix) {
					fieldName = strings.TrimSuffix(fieldName, suffix)
				} else {
					for _, otherLanguage := range languages {
						if strings.HasSuffix(fieldName, "-"+otherLanguage) {
							continue FieldLoop
						}
					}
				}
				languageData[strings.TrimSuffix(fieldName, suffix)] = data
			}
			fmt.Println("generating doc", i, id, lang, title)
			spew.Dump(languageData)
			targetFilename := filepath.Join(outDir, title)
			os.Remove(targetFilename)
			file, errOpen := os.OpenFile(targetFilename, os.O_RDWR|os.O_CREATE, 0644)
			must("could not open file", errOpen)
			errExecute := tpl.Execute(file, languageData)
			must("template exec", errExecute)
			must("could not close file", file.Close())
		}
	}
}

func main() {
	flagLangs := flag.String("languages", "de,en", "comma separated list of languages")
	flagCSV := flag.String("csv", "", "csv url like https://csv.com/doc.csv")
	flagOutDir := flag.String("out", "", "output directory")
	flagTemplate := flag.String("template", "", "path/to/template.html")
	flagAssetDir := flag.String("asset-dir", "", "path/to/asset/dir")
	flag.Parse()

	if *flagTemplate == "" {
		flag.Usage()
		must("arg missing", errors.New("no template given"))
	}

	if *flagCSV == "" {
		flag.Usage()
		must("arg missing", errors.New("no csv url given"))
	}

	if *flagOutDir == "" {
		flag.Usage()
		must("arg missing", errors.New("no output dir given"))
	}

	outDir := *flagOutDir
	var source io.ReadCloser

	if strings.HasPrefix(*flagCSV, "http") {
		resp, errGet := http.Get(*flagCSV)
		must("csv download failed "+*flagCSV, errGet)
		if resp.StatusCode != http.StatusOK {
			must("donwload failed", errors.New("download failed: "+resp.Status))
		}
		source = resp.Body
	} else {
		fileSource, errOpen := os.Open(*flagCSV)
		must("could not open csv file", errOpen)
		source = fileSource
	}
	defer source.Close()

	td := readData(source)

	tpl, errTPL := template.New(filepath.Base(*flagTemplate)).Funcs(getTemplateFuncs(*flagAssetDir)).ParseFiles(*flagTemplate)
	must("template parsing", errTPL)

	languages := strings.Split(*flagLangs, ",")

	renderData(td, languages, outDir, tpl)

}
