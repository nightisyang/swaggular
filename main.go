package main

import (
	Apis "angular-service-builder/pkg/apis"
	Dtos "angular-service-builder/pkg/dtos"
	Types "angular-service-builder/pkg/types"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ProcessedRefs is a global map to prevent infinite recursion.
var processedRefs = map[string]bool{}

// Precomputed DTOs and API list
var (
	precomputedDTOs    map[string]string
	precomputedAPIList []Types.APIWithDTO
)

// InitializeData initializes the data for the application.
func InitializeData() {
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error determining executable path: %v", err)
	}
	executableDir := filepath.Dir(executablePath)

	swaggerPath := filepath.Join(executableDir, "swagger.json")
	data, err := ioutil.ReadFile(swaggerPath)
	if err != nil {
		log.Fatalf("Error reading swagger.json: %v", err)
	}

	var openAPI Types.OpenAPI
	err = json.Unmarshal(data, &openAPI)
	if err != nil {
		log.Fatalf("Error unmarshalling swagger.json: %v", err)
	}

	precomputedDTOs = make(map[string]string)
	for name, schema := range openAPI.Components.Schemas {
		Dtos.GenerateTypeScriptInterface(name, schema)
	}
	fmt.Println("Finished pre-generating DTOs")

	precomputedAPIList = Apis.GenerateAPIList(openAPI)
	fmt.Println("Finished pre-generating API list")

	writeDTOsToFile("dtos.txt")
}

// WriteDTOsToFile writes DTOs to a file.
func writeDTOsToFile(filename string) {
	fmt.Println("Attempting to create file:", filename)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	for key, dto := range Dtos.DTOMap {
		fmt.Println("Writing DTO:", key)
		_, err := file.WriteString(dto)
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			return
		}
	}
	fmt.Printf("DTOs successfully written to %s\n", filename)
}

// ServeIndex serves the main page.
func serveIndex(w http.ResponseWriter, r *http.Request) {
	tmplData := Types.TemplateData{APIList: precomputedAPIList}

	executablePath, executableErr := os.Executable()
	if executableErr != nil {
		log.Fatalf("Error determining executable path: %v", executableErr)
	}
	executableDir := filepath.Dir(executablePath)

	indexHtmlFilePath := filepath.Join(executableDir, "templates", "index.html")
	tmpl := template.Must(template.ParseFiles(indexHtmlFilePath))

	serveError := tmpl.Execute(w, tmplData)
	if serveError != nil {
		http.Error(w, serveError.Error(), http.StatusInternalServerError)
	}
}

// ServeAPIDetails serves the details of a specific API.
func serveAPIDetails(w http.ResponseWriter, r *http.Request) {
	apiName := r.URL.Query().Get("api")

	var selectedAPI *Types.APIWithDTO
	for _, api := range precomputedAPIList {
		if api.FunctionName == apiName {
			selectedAPI = &api
			break
		}
	}

	if selectedAPI != nil {
		tmpl := template.Must(template.ParseFiles(filepath.Join("templates", "api_details.html")))

		dtos := collectAllDTOs(selectedAPI.ResponseType)
		if selectedAPI.PayloadType != "" {
			dtos = append(dtos, collectAllDTOs(selectedAPI.PayloadType)...)
		}

		apiDetail := struct {
			API  *Types.APIWithDTO
			DTOs []string
		}{
			API:  selectedAPI,
			DTOs: dtos,
		}

		err := tmpl.Execute(w, apiDetail)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "API not found", http.StatusNotFound)
	}
}

// CollectAllDTOs collects all DTOs related to a given name.
func collectAllDTOs(name string) []string {
	var collectedDTOs []string
	baseName := strings.TrimSuffix(name, "[]")
	processed := map[string]bool{}

	var collect func(name string)
	collect = func(name string) {
		if _, exists := processed[name]; exists {
			return
		}
		processed[name] = true

		if dto, exists := Dtos.DTOMap[name]; exists {
			collectedDTOs = append(collectedDTOs, dto)

			for key := range Dtos.DTOMap {
				if strings.Contains(dto, key) && key != name {
					collect(key)
				}
			}
		}
	}

	collect(baseName)
	return collectedDTOs
}

// Main function to start the server and serve the pages.
func main() {
	InitializeData()

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/api-detail", serveAPIDetails)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
