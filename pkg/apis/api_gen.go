package Apis

import (
	Dtos "angular-service-builder/pkg/dtos"
	Types "angular-service-builder/pkg/types"
	"fmt"
	"strings"
)

// GenerateAPIList generates a list of API methods with DTOs.
func GenerateAPIList(api Types.OpenAPI) []Types.APIWithDTO {
	var apiList []Types.APIWithDTO

	for path, operations := range api.Paths {
		for method, operation := range operations {
			functionName := CreateFunctionName(operation.OperationID)

			var pathParams []string
			var queryParamInterfaceBuilder strings.Builder
			var payloadType string
			hasActualQueryParams := false

			queryParamInterfaceBuilder.WriteString(fmt.Sprintf("export interface %sQueryParams {\n", functionName))

			for _, param := range operation.Parameters {
				camelCaseName := Dtos.ToCamelCase(param.Name)

				if param.In == "path" {
					pathParams = append(pathParams, fmt.Sprintf("%s: %s", camelCaseName, Dtos.MapType(param.Schema)))
				} else if param.In == "query" {
					hasActualQueryParams = true
					paramLine := fmt.Sprintf("  %s%s: %s;", camelCaseName, Dtos.OptionalSuffix(param.Required), Dtos.MapType(param.Schema))
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

			if operation.RequestBody != nil {
				for _, mediaType := range operation.RequestBody.Content {
					if mediaType.Schema.Ref != "" {
						refName := strings.TrimPrefix(mediaType.Schema.Ref, "#/components/schemas/")
						payloadType = refName

						if schema, exists := api.Components.Schemas[refName]; exists {
							Dtos.GenerateTypeScriptInterface(refName, schema)
						}
						break
					}
				}
			}

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

			for _, param := range operation.Parameters {
				if param.In == "path" {
					path = strings.Replace(path, fmt.Sprintf("{%s}", param.Name), fmt.Sprintf("${%s}", Dtos.ToCamelCase(param.Name)), -1)
				}
			}

			apiList = append(apiList, Types.APIWithDTO{
				FunctionName:        functionName,
				Parameters:          parameters,
				QueryParamInterface: queryParamInterface,
				ResponseType:        responseType,
				PayloadType:         payloadType,
				Path:                path,
				HasQueryParams:      hasQueryParams,
				HttpMethod:          method,
			})
		}
	}

	return apiList
}

// CreateFunctionName converts an operation ID to camel case.
func CreateFunctionName(operationID string) string {
	return Dtos.ToCamelCase(operationID)
}
