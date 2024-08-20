package Dtos

import (
	Types "angular-service-builder/pkg/types"
	"fmt"
	"strings"
)

// DTOMap stores all generated DTOs, keyed by their names.
var DTOMap = make(map[string]string)

// MapType maps OpenAPI schema types to TypeScript types.
func MapType(schema Types.Schema) string {
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
				return MapType(*schema.Items) + "[]"
			}
		}
	} else if schema.Ref != "" {
		return schema.Ref[strings.LastIndex(schema.Ref, "/")+1:]
	}
	return "any"
}

// GenerateTypeScriptInterface generates TypeScript interfaces for DTOs.
func GenerateTypeScriptInterface(name string, schema Types.Schema) {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("export interface %s {\n", name))
	for propName, propSchema := range schema.Properties {
		camelCasePropName := ToCamelCase(propName) // Convert to camelCase
		nullableSuffix := " | null"

		if propSchema.Type == "object" {
			nestedInterfaceName := propName + "DTO"
			GenerateTypeScriptInterface(nestedInterfaceName, propSchema)
			builder.WriteString(fmt.Sprintf("  %s: %s%s;\n", camelCasePropName, nestedInterfaceName, nullableSuffix))
		} else if propSchema.Type == "array" {
			if propSchema.Items != nil {
				itemType := MapType(*propSchema.Items)
				if propSchema.Items.Type == "object" {
					nestedInterfaceName := propName + "DTO"
					GenerateTypeScriptInterface(nestedInterfaceName, *propSchema.Items)
					itemType = nestedInterfaceName
				}
				builder.WriteString(fmt.Sprintf("  %s: %s[]%s;\n", camelCasePropName, itemType, nullableSuffix))
			}
		} else {
			builder.WriteString(fmt.Sprintf("  %s: %s%s;\n", camelCasePropName, MapType(propSchema), nullableSuffix))
		}
	}
	builder.WriteString("}\n\n")
	DTOMap[name] = builder.String()
}
