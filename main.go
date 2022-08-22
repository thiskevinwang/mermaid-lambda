package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func mmdcSvg(w http.ResponseWriter, r *http.Request) {
	// reject if not POST or GET
	if r.Method != "POST" && r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// get query params
	query := r.URL.Query()
	theme := query.Get("theme")
	// base64 encoded input in qsparam
	var b64input string
	var input string

	switch r.Method {
	case "POST":
		// get request body
		defer r.Body.Close()
		// reject empty body
		body, _ := ioutil.ReadAll(r.Body)
		if len(body) == 0 {
			http.Error(w, "No body", http.StatusBadRequest)
			return
		}
		input = string(body)
	case "GET":
		b64input = query.Get("input")
		// reject empty input
		if len(b64input) == 0 {
			http.Error(w, "No input", http.StatusBadRequest)
			return
		}

		// "+" in querystring becomes whitespace.
		// We should replace it back to "+" to ensure it can be base64 decoded.
		b64input = strings.Replace(b64input, " ", "+", -1)

		// decode base64 input
		if b64decoded, err := base64.StdEncoding.DecodeString(b64input); err != nil {
			fmt.Println("b64input", b64input)
			fmt.Println("Error decoding base64 input:", err)
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		} else {
			input = string(b64decoded)
		}
	}

	// The path to the mermaid CLI executable
	//
	// WARN: when executing the mmdc bin, `node` is required!
	var binPath string
	if dirname, err := os.Getwd(); err != nil {
		log.Fatal("failed to get working dir", err)
	} else {
		binPath = filepath.Join(dirname, "./node_modules/.bin/mmdc")
	}

	// /tmp is AWS Lambda's writeable tmp dir
	var outputFile string = "/tmp/output.png"

	// Base Mermaid CLI command
	cmd := exec.Command(binPath, "-o", outputFile, "-p", "puppeteer-config.json", "-t", theme, "-b", "transparent")

	// Pipe stdin to the mermaid CLI:
	// How-to in Go: https://stackoverflow.com/a/49901167/9823455
	cmd.Stdin = strings.NewReader(input)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		log.Fatal("command failed: ", err)
	}

	svg, err := os.ReadFile(outputFile)
	if err != nil {
		log.Fatal("failed to read outfile: ", err)
	}

	res := string(svg)

	// Set header for CORS here
	//
	// When Hard-refreshing on Chrome, the initial requests to CloudFront don't
	// get the correct response headers, and result in a CORs error.
	w.Header().Set("access-control-allow-origin", "*")

	// w.Header().Set("content-type", "image/svg+xml")
	w.Header().Set("content-type", "text/plain")

	w.Write([]byte(res))
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello World!")
	})
	http.HandleFunc("/generate", mmdcSvg)

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
