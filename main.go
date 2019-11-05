package main

import (
	"fmt"
	"io"
	"strings"
	"log"
	"net/http"
	"encoding/csv"
	"strconv"

	"github.com/wcharczuk/go-chart" //exposes "chart"
)

const MainPageContent string = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
</head>
<body>
%s
<form action="./" method="post" enctype="multipart/form-data">
<p><label>ファイル選択:<br><input type="file" name="csv"></label></p>
<input type="submit" value="Upload">
</form>
<a href="./">main</a>
</body>
</html>
`

const DefaultSvg string = `<svg
   xmlns:dc="http://purl.org/dc/elements/1.1/"
   xmlns:cc="http://creativecommons.org/ns#"
   xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
   xmlns:svg="http://www.w3.org/2000/svg"
   xmlns="http://www.w3.org/2000/svg"
   id="svg8"
   version="1.1"
   viewBox="0 0 105.83334 79.375"
   height="300"
   width="400.00003">
  <defs
     id="defs2" />
  <metadata
     id="metadata5">
    <rdf:RDF>
      <cc:Work
         rdf:about="">
        <dc:format>image/svg+xml</dc:format>
        <dc:type
           rdf:resource="http://purl.org/dc/dcmitype/StillImage" />
        <dc:title></dc:title>
      </cc:Work>
    </rdf:RDF>
  </metadata>
  <g
     transform="translate(-71.815475,-166.30032)"
     id="layer1">
    <rect
       style="fill:#ffffff;stroke-width:0.26458332"
       y="166.30032"
       x="71.815475"
       height="79.375"
       width="105.83334"
       id="rect3713" />
    <text
       id="text3717"
       y="209.95399"
       x="105.15453"
       style="font-style:normal;font-weight:normal;font-size:10.58333302px;line-height:1.25;font-family:sans-serif;letter-spacing:0px;word-spacing:0px;fill:#000000;fill-opacity:1;stroke:none;stroke-width:0.26458332"
       xml:space="preserve"><tspan
         style="stroke-width:0.26458332"
         y="209.95399"
         x="105.15453"
         id="tspan3715">Default</tspan></text>
  </g>
</svg>`

func ServeMainPage(w http.ResponseWriter, r *http.Request, svg string) {
	if svg == "" {
		svg = DefaultSvg
	}
	fmt.Fprintf(w, MainPageContent, svg)
}

func Serve(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	svg := ""
	switch r.Method {
	case "GET":
		ServeMainPage(w, r, svg)
	case "POST":
		csvFile, _, err := r.FormFile("csv")
		if err != nil {
			ServeMainPage(w, r, svg)
			break
		}
		svg = SvgPlotOfCsv(csvFile)
		ServeMainPage(w, r, svg)
	default:
		fmt.Fprintf(w, "Not Implemented")
	}
}

func indexSeries(data []float64) []float64 {
	index := make([]float64, len(data))
	for i := range index {
		index[i] = float64(i)
	}
	return index
}

func SvgPlotOfCsv(csvFile io.Reader) string {
	strMat, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		return ""
	}
	if len(strMat) == 0 {
		return ""
	}
	rows := len(strMat)
	cols := len(strMat[0])
	columns := make([][]float64, cols)
	for c := range columns {
		columns[c] = make([]float64, rows)
		for r := range columns[c] {
			if columns[c][r], err = strconv.ParseFloat(strMat[r][c], 64); err != nil {
				return ""
			}
		}
	}
	if len(columns) == 0 {
		return ""
	}
	columns = append(columns[0:1], columns[0:]...)
	columns[0] = indexSeries(columns[1])
	var svg strings.Builder
	graph := chart.Chart{
		Series: make([]chart.Series, len(columns) - 1),
	}
	for i := range graph.Series {
		graph.Series[i] = chart.ContinuousSeries{XValues: columns[0], YValues: columns[i + 1]}
	}
	graph.Render(chart.SVG, &svg)
	return svg.String()
}

func main() {
	http.HandleFunc("/", Serve)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
