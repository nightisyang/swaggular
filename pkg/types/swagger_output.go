package Types

// Structs to hold API information for rendering.
type APIWithDTO struct {
	FunctionName        string
	Parameters          string
	QueryParamInterface string
	ResponseType        string
	PayloadType         string
	Path                string
	HasQueryParams      bool
	HttpMethod          string
}

type TemplateData struct {
	APIList []APIWithDTO
}
