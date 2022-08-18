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
		// base64 encode body for cache key
		b64input = base64.StdEncoding.EncodeToString(body)
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

		// WARN this may be redundant as we need to URL encode a request to a Lambda Function URL anyways
		// Presence of "=" in the URL returns 400, {"message": null}

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

	key := fmt.Sprintf("%s::%s", b64input, theme)
	fmt.Println("hash key:", key)

	mmdcPath := "./node_modules/.bin/mmdc"
	__dirname, err := os.Getwd()
	if err != nil {
		log.Fatal("failed to get working dir", err)
	}

	binPath := filepath.Join(__dirname, mmdcPath)

	// WARN — when executing the mmdc bin, `node` is required!

	// How to pipe string to command?
	// https://stackoverflow.com/a/49901167/9823455

	var outputFile string = "/tmp/output.svg"
	var cmd *exec.Cmd

	cmd = exec.Command(binPath, "-o", outputFile, "-p", "puppeteer-config.json", "-t", theme, "-b", "transparent")

	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	fmt.Println("Executing:", cmd.String())
	if err := cmd.Run(); err != nil {
		log.Fatal("command failed: ", err)
	}
	fmt.Println("Command finished")

	fmt.Println("Sending file to client")
	svg, err := os.ReadFile(outputFile)
	if err != nil {
		log.Fatal("failed to read outfile: ", err)
	}

	res := string(svg)

	// if GET, set cache headers for 1 year
	// if r.Method == "GET" {
	// 	w.Header().Set("Cache-Control", "public, max-age=31536000")
	// }
	w.Header().Set("Content-Type", "text/plain")

	// set this header for CORS here
	// This addresses an error:
	// When Hard-refreshing on Chrome, the initial requests to CloudFront don't
	// get the correct response headers, and result in a CORs error.
	w.Header().Set("access-control-allow-origin", "*")
	w.Header().Set("content-type", "image/svg+xml")
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
