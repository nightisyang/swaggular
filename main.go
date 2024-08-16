package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Define the Schema struct to match the structure of the OpenAPI JSON.
type Schema struct {
	Type        string            `json:"type"`
	Format      string            `json:"format,omitempty"`
	Ref         string            `json:"$ref,omitempty"`
	Items       *Schema           `json:"items,omitempty"`
	Properties  map[string]Schema `json:"properties,omitempty"`
	Required    []string          `json:"required,omitempty"`
	Description string            `json:"description,omitempty"`
}

// Define the other structs for OpenAPI components.
type OpenAPI struct {
	Paths      map[string]map[string]Operation `json:"paths"`
	Components Components                      `json:"components"`
}

type Operation struct {
	OperationID string              `json:"operationId"`
	Parameters  []Parameter         `json:"parameters"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Content     map[string]MediaType `json:"content"`
	Required    bool                 `json:"required,omitempty"`
}

type Parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
	Schema   Schema `json:"schema"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Components struct {
	Schemas map[string]Schema `json:"schemas"`
}

// ProcessedRefs is a global map to prevent infinite recursion.
var processedRefs = map[string]bool{}

// DTOMap stores all generated DTOs, keyed by their names.
var DTOMap = make(map[string]string)

// Function to map OpenAPI schema types to TypeScript types.
func mapType(schema Schema) string {
	if schema.Type != "" {
		switch schema.Type {
		case "string":
			return "string"
		case "integer":
			return "number"
		case "boolean":
			return "boolean"
		case "array":
			if schema.Items != nil {
				return mapType(*schema.Items) + "[]"
			}
		}
	} else if schema.Ref != "" {
		if processedRefs[schema.Ref] {
			return schema.Ref[strings.LastIndex(schema.Ref, "/")+1:]
		}
		processedRefs[schema.Ref] = true
		return schema.Ref[strings.LastIndex(schema.Ref, "/")+1:]
	}
	return "any"
}

// Function to pre-generate all TypeScript interfaces for DTOs in the components.
func preGenerateDTOs(components Components) {
	for name, schema := range components.Schemas {
		generateTypeScriptInterface(name, schema)
	}
}

// Function to generate TypeScript interfaces for DTOs, ensuring no duplicates.
func generateTypeScriptInterface(name string, schema Schema) {
	if _, exists := DTOMap[name]; exists {
		return
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("export interface %s {\n", name))
	for propName, propSchema := range schema.Properties {
		required := ""
		if contains(schema.Required, propName) {
			required = " // required"
		}

		// Check if the property is a complex type (object or array of objects)
		if propSchema.Type == "object" {
			nestedInterfaceName := propName + "DTO"
			generateTypeScriptInterface(nestedInterfaceName, propSchema)
			builder.WriteString(fmt.Sprintf("  %s: %s;%s\n", propName, nestedInterfaceName, required))
		} else if propSchema.Type == "array" {
			if propSchema.Items != nil {
				itemType := mapType(*propSchema.Items)
				if propSchema.Items.Type == "object" {
					nestedInterfaceName := propName + "DTO"
					generateTypeScriptInterface(nestedInterfaceName, *propSchema.Items)
					itemType = nestedInterfaceName
				}
				builder.WriteString(fmt.Sprintf("  %s: %s[];%s\n", propName, itemType, required))
			}
		} else {
			// Primitive types or arrays of primitive types
			builder.WriteString(fmt.Sprintf("  %s: %s;%s\n", propName, mapType(propSchema), required))
		}
	}
	builder.WriteString("}\n\n")
	DTOMap[name] = builder.String()
}

// Collect all nested DTOs based on the main DTO
func collectAllDTOs(name string) []string {
	var collectedDTOs []string

	// Track processed DTOs to avoid duplication
	processed := map[string]bool{}

	var collect func(name string)
	collect = func(name string) {
		if _, exists := processed[name]; exists {
			return
		}
		processed[name] = true

		if dto, exists := DTOMap[name]; exists {
			collectedDTOs = append(collectedDTOs, dto)

			// Recursively collect nested DTOs
			for key := range DTOMap {
				if strings.Contains(dto, key) && key != name {
					collect(key)
				}
			}
		}
	}

	collect(name)
	return collectedDTOs
}

// Function to generate the list of API methods, including query params, payload, and response handling.
func generateAPIList(api OpenAPI) []APIWithDTO {
	var apiList []APIWithDTO

	for path, operations := range api.Paths {
		for method, operation := range operations {
			functionName := createFunctionName(method, operation.OperationID)

			var pathParams []string
			var queryParamInterfaceBuilder strings.Builder
			var payloadType string
			hasActualQueryParams := false

			// Build the query parameter interface
			queryParamInterfaceBuilder.WriteString(fmt.Sprintf("export interface %sQueryParams {\n", functionName))

			for _, param := range operation.Parameters {
				camelCaseName := toCamelCase(param.Name)

				if param.In == "path" {
					pathParams = append(pathParams, fmt.Sprintf("%s: %s", camelCaseName, mapType(param.Schema)))
				} else if param.In == "query" {
					hasActualQueryParams = true
					paramLine := fmt.Sprintf("  %s%s: %s;", camelCaseName, optionalSuffix(param.Required), mapType(param.Schema))
					queryParamInterfaceBuilder.WriteString(paramLine + "\n")
				}
			}

			queryParamInterfaceBuilder.WriteString("}\n")
			queryParamInterface := queryParamInterfaceBuilder.String()

			parameters := strings.Join(pathParams, ", ")
			hasQueryParams := hasActualQueryParams

			if hasQueryParams {
				if parameters != "" {
					parameters += ", "
				}
				parameters += "queryParams: " + functionName + "QueryParams"
			}

			// Handle request body (payload)
			if operation.RequestBody != nil {
				for _, mediaType := range operation.RequestBody.Content {
					if mediaType.Schema.Ref != "" {
						refName := strings.TrimPrefix(mediaType.Schema.Ref, "#/components/schemas/")
						payloadType = refName

						// Generate the interface for the payload
						if schema, exists := api.Components.Schemas[refName]; exists {
							generateTypeScriptInterface(refName, schema)
						}
						break
					}
				}
			}

			// Determine the response type based on the 200 response
			responseType := "void"
			if resp, ok := operation.Responses["200"]; ok {
				for _, mediaType := range resp.Content {
					if mediaType.Schema.Type == "array" && mediaType.Schema.Items != nil && mediaType.Schema.Items.Ref != "" {
						refName := strings.TrimPrefix(mediaType.Schema.Items.Ref, "#/components/schemas/")
						responseType = fmt.Sprintf("%s[]", refName)
					} else if mediaType.Schema.Ref != "" {
						refName := strings.TrimPrefix(mediaType.Schema.Ref, "#/components/schemas/")
						responseType = refName
					}
				}
			}

			// Update the path to include dynamic parameters
			for _, param := range operation.Parameters {
				if param.In == "path" {
					path = strings.Replace(path, fmt.Sprintf("{%s}", param.Name), fmt.Sprintf("${%s}", toCamelCase(param.Name)), -1)
				}
			}

			apiList = append(apiList, APIWithDTO{
				FunctionName:        functionName,
				Parameters:          parameters,
				QueryParamInterface: queryParamInterface,
				ResponseType:        responseType,
				PayloadType:         payloadType,
				Path:                path,
				HasQueryParams:      hasQueryParams,
			})
		}
	}

	return apiList
}

// Structs to hold API information for rendering.
type APIWithDTO struct {
	FunctionName        string
	Parameters          string
	QueryParamInterface string
	ResponseType        string
	PayloadType         string
	Path                string
	HasQueryParams      bool
}

type TemplateData struct {
	APIList []APIWithDTO
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("swagger.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var openAPI OpenAPI
	err = json.Unmarshal(data, &openAPI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Pre-generate all DTOs
	preGenerateDTOs(openAPI.Components)

	// Generate the API list with the updated query parameters and other details
	apiList := generateAPIList(openAPI)
	tmplData := TemplateData{APIList: apiList}

	// Parse and execute the template with the generated API list
	tmpl := template.Must(template.New("index").Parse(indexTemplate))
	err = tmpl.Execute(w, tmplData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func serveAPIDetails(w http.ResponseWriter, r *http.Request) {
	apiName := r.URL.Query().Get("api")

	data, err := ioutil.ReadFile("swagger.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var openAPI OpenAPI
	err = json.Unmarshal(data, &openAPI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiList := generateAPIList(openAPI)
	var selectedAPI *APIWithDTO
	for _, api := range apiList {
		if api.FunctionName == apiName {
			selectedAPI = &api
			break
		}
	}

	if selectedAPI != nil {
		tmpl := template.Must(template.New("apiDetail").Parse(apiDetailTemplate))

		// Collect all DTOs related to the selected API's response and payload types
		dtos := collectAllDTOs(selectedAPI.ResponseType)
		if selectedAPI.PayloadType != "" {
			dtos = append(dtos, collectAllDTOs(selectedAPI.PayloadType)...)
		}

		apiDetail := struct {
			API  *APIWithDTO
			DTOs []string
		}{
			API:  selectedAPI,
			DTOs: dtos,
		}

		err = tmpl.Execute(w, apiDetail)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "API not found", http.StatusNotFound)
	}
}

func toCamelCase(input string) string {
	parts := strings.Split(input, "_")
	for i := range parts {
		if i > 0 {
			parts[i] = strings.Title(parts[i])
		}
	}
	return strings.Join(parts, "")
}

// optionalSuffix adds a question mark for optional query parameters.
func optionalSuffix(isRequired bool) string {
	if isRequired {
		return ""
	}
	return "?"
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func createFunctionName(method, operationID string) string {
	return strings.ToLower(method) + strings.Title(operationID)
}

// Main function to start the server and serve the pages.
func main() {
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/api-detail", serveAPIDetails)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// HTML template for the main page listing the APIs.
const indexTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API List</title>
    <script src="https://unpkg.com/htmx.org@1.4.0"></script>
</head>
<body>
    <h1>Angular Service Code and DTO Interfaces</h1>
    {{range .APIList}}
        <div>
            <button hx-get="/api-detail?api={{.FunctionName}}" hx-target="#details-{{.FunctionName}}" hx-swap="outerHTML">
                {{.FunctionName}}
            </button>
            <div id="details-{{.FunctionName}}" style="margin-left: 20px;">
                <!-- Details will be loaded here by htmx -->
            </div>
        </div>
        <hr>
    {{end}}
</body>
</html>
`

// HTML template for the API details (method, payload DTOs, and response DTOs).
const apiDetailTemplate = `
<pre>
  {{if .API.HasQueryParams}}
    {{.API.QueryParamInterface}}
  {{end}}

  {{.API.FunctionName}}({{if .API.Parameters}}{{.API.Parameters}}, {{end}}{{if .API.PayloadType}}payload: {{.API.PayloadType}}{{end}}{{if .API.HasQueryParams}}, queryParams: {{.API.FunctionName}}QueryParams{{end}}): Observable<{{.API.ResponseType}}> {
    {{if .API.HasQueryParams}}
      let params = httpParamBuilder(params) 
    {{end}}
    
    return this.http.{{if .API.PayloadType}}put<{{.API.ResponseType}}>("{{.API.Path}}", payload {{if .API.HasQueryParams}}, { params } {{end}} ){{else}}get<{{.API.ResponseType}}>("{{.API.Path}}", { params }){{end}};
  }
</pre>
{{range .DTOs}}
<pre>{{.}}</pre>
{{else}}
<pre>No DTOs available.</pre>
{{end}}
<hr>
`
