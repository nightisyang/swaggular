package Types

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
