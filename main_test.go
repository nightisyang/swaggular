package main

import (
	"testing"
)

func TestGenerateTypeScriptInterface(t *testing.T) {
	DTOMap = make(map[string]string) // Reset DTOMap before each test

	schema := Schema{
		Type: "object",
		Properties: map[string]Schema{
			"transactionChannelSettingDropdowns": {Type: "array", Items: &Schema{Type: "object", Properties: map[string]Schema{
				"id":   {Type: "string"},
				"name": {Type: "string"},
			}}},
			"issuerSettingDropdowns": {Type: "array", Items: &Schema{Type: "object", Properties: map[string]Schema{
				"id":   {Type: "string"},
				"name": {Type: "string"},
			}}},
		},
	}

	generateTypeScriptInterface("GetSPMSettingDropdownOutputDTO", schema)

	if _, exists := DTOMap["GetSPMSettingDropdownOutputDTO"]; !exists {
		t.Errorf("Expected GetSPMSettingDropdownOutputDTO to be generated")
	}

	if _, exists := DTOMap["transactionChannelSettingDropdownsDTO"]; !exists {
		t.Errorf("Expected transactionChannelSettingDropdownsDTO to be generated")
	}

	if _, exists := DTOMap["issuerSettingDropdownsDTO"]; !exists {
		t.Errorf("Expected issuerSettingDropdownsDTO to be generated")
	}
}

func TestCollectAllDTOs(t *testing.T) {
	// Setup DTOMap with sample data
	DTOMap = map[string]string{
		"GetSPMSettingDropdownOutputDTO": `
export interface GetSPMSettingDropdownOutputDTO {
  transactionChannelSettingDropdowns: TransactionChannelSettingDropdown[];
  issuerSettingDropdowns: IssuerSettingDropdown[];
}
`,
		"TransactionChannelSettingDropdown": `
export interface TransactionChannelSettingDropdown {
  id: string;
  name: string;
}
`,
		"IssuerSettingDropdown": `
export interface IssuerSettingDropdown {
  id: string;
  name: string;
}
`,
	}

	collectedDTOs := collectAllDTOs("GetSPMSettingDropdownOutputDTO")

	expectedDTOs := []string{
		DTOMap["GetSPMSettingDropdownOutputDTO"],
		DTOMap["TransactionChannelSettingDropdown"],
		DTOMap["IssuerSettingDropdown"],
	}

	for _, expectedDTO := range expectedDTOs {
		if !containsString(collectedDTOs, expectedDTO) {
			t.Errorf("Expected DTO not found: %s", expectedDTO)
		}
	}
}

func containsString(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func TestGenerateTypeScriptInterfaceForGetParameterSettings(t *testing.T) {
	// Reset DTOMap before each test
	DTOMap = make(map[string]string)

	// Define the schema for GetParameterSettingsOutputDTOPagedResult
	schema := Schema{
		Type: "object",
		Properties: map[string]Schema{
			"currentPage":    {Type: "number"},
			"pageCount":      {Type: "number"},
			"pageSize":       {Type: "number"},
			"rowCount":       {Type: "number"},
			"firstRowOnPage": {Type: "number"},
			"lastRowOnPage":  {Type: "number"},
			"results": {Type: "array", Items: &Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id":                {Type: "number"},
					"name":              {Type: "string"},
					"settingType":       {Type: "string"},
					"settingValue":      {Type: "string"},
					"settingStatusCode": {Type: "number"},
				},
			}},
		},
	}

	// Generate the TypeScript interface for the schema
	generateTypeScriptInterface("GetParameterSettingsOutputDTOPagedResult", schema)

	// Check that GetParameterSettingsOutputDTOPagedResult is generated
	if _, exists := DTOMap["GetParameterSettingsOutputDTOPagedResult"]; !exists {
		t.Errorf("Expected GetParameterSettingsOutputDTOPagedResult to be generated")
	}

	// Check that GetParameterSettingsOutputDTO (nested DTO) is generated
	if _, exists := DTOMap["resultsDTO"]; !exists {
		t.Errorf("Expected GetParameterSettingsOutputDTO to be generated as resultsDTO")
	}
}

// Test for generating and storing all related DTOs
func TestGenerateTypeScriptInterfaceForTransactionLimitSettingFilters(t *testing.T) {
	// Reset DTOMap before each test
	DTOMap = make(map[string]string)

	// Define the schema for GetTransactionLimitSettingFiltersOutputDTO
	schema := Schema{
		Type: "object",
		Properties: map[string]Schema{
			"supplyPartnerDropdowns": {Type: "array", Items: &Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id":   {Type: "string"},
					"name": {Type: "string"},
				},
			}},
			"supplyPartnerWalletDropdowns": {Type: "array", Items: &Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id":   {Type: "string"},
					"name": {Type: "string"},
				},
			}},
		},
	}

	// Generate the TypeScript interface for the schema
	generateTypeScriptInterface("GetTransactionLimitSettingFiltersOutputDTO", schema)

	// Check that the main DTO is generated
	if _, exists := DTOMap["GetTransactionLimitSettingFiltersOutputDTO"]; !exists {
		t.Errorf("Expected GetTransactionLimitSettingFiltersOutputDTO to be generated")
	}

	// Check that the nested DTOs are generated
	if _, exists := DTOMap["supplyPartnerDropdownsDTO"]; !exists {
		t.Errorf("Expected TLSupplyPartnerDropdownDTO to be generated as supplyPartnerDropdownsDTO")
	}

	if _, exists := DTOMap["supplyPartnerWalletDropdownsDTO"]; !exists {
		t.Errorf("Expected TLSupplyPartnerWalletDropdownDTO to be generated as supplyPartnerWalletDropdownsDTO")
	}
}
