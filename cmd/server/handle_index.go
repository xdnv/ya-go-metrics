package main

import (
	"fmt"
	"io"

	"net/http"
)

// const metricsHTML = `
//  	<h1>{{.PageTitle}}</h1>
// 	 <style>
// 	 table, td, th {
// 	   border: 1px solid black;
// 	   border-spacing: 0px;
// 	 }
// 	 </style>
//  	<table>
// 		{{range .Metrics}}
// 			{{if .Header}}
// 				<tr><th>Metric</th><th>Value</th></tr>
// 			{{else}}
// 				<tr><td>{{.Title}}</td><td style=\"text-align: right;\">{{.Value}}</td></tr>
// 			{{end}}
// 		{{end}}
// 	</table>
// 	`

// type MetricEntry struct {
// 	Title  string
// 	Value  string
// 	Header bool
// }

// type MetricPageData struct {
// 	PageTitle string
// 	Metrics   []MetricEntry
// }

// func index_t(w http.ResponseWriter, r *http.Request) {

// 	//check for malformed requests
// 	if r.URL.Path != "/" {
// 		http.NotFound(w, r)
// 		return
// 	}

// 	// set correct datatype in header
// 	w.Header().Set("Content-Type", "text/html; charset=utf-8")
// 	w.WriteHeader(http.StatusOK)

// 	data := new(MetricPageData)

// 	data.PageTitle = "Current values"
// 	data.Metrics = append(data.Metrics, MetricEntry{"Metric", "Value", true})

// 	for _, key := range sortedKeys(storage.Metrics) {
// 		data.Metrics = append(data.Metrics, MetricEntry{key, fmt.Sprintf("%v", storage.Metrics[key].(Metric).GetValue()), false})
// 	}

// 	tmpl := template.Must(template.New("").Parse(metricsHTML))
// 	tmpl.Execute(w, data)
// }

const indexPageTpl = `<html>
	<head>
		<title>%s</title>
		<style>
	  		table, td, th {
	    		border: 1px solid black;
	    		border-spacing: 0px;
	  		}
	  		td, th {
	    		padding: 5px;
	  		}
		</style>
	</head>
	<body>
	   	%s
	</body>
	</html>`

const indexTableTpl = "<table>%s</table>"
const indexTableHeaderTpl = "<tr><th>%s</th><th>%v</th></tr>"
const indexTableRowTpl = "<tr><td>%s</td><td style=\"text-align: right;\">%v</td></tr>"

func index(w http.ResponseWriter, r *http.Request) {
	//check for malformed requests - only exact root path accepted
	//Important: covered by tests, removal will bring tests to fail
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// set correct data type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	htmlBody := fmt.Sprintf(indexTableHeaderTpl, "Metric", "Value")

	for _, key := range sortKeys(storage.Metrics) {
		htmlBody += fmt.Sprintf(indexTableRowTpl, key, storage.Metrics[key].GetValue())
	}
	htmlBody = fmt.Sprintf(indexTableTpl, htmlBody)

	io.WriteString(w, fmt.Sprintf(indexPageTpl, "Metrics", htmlBody))
}
