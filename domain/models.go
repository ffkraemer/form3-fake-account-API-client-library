package domain

import "time"

//region COMMON MODELS

type Attributes struct {
	Country          string   `json:"country,omitempty"`
	BaseCurrency     string   `json:"base_currency,omitempty"`
	BankID           string   `json:"bank_id,omitempty"`
	BankIDCode       string   `json:"bank_id_code,omitempty"`
	Bic              string   `json:"bic,omitempty"`
	Name             []string `json:"name,omitempty"`
	AlternativeNames []string `json:"alternative_names,omitempty"`
	UserDefinedData  []struct {
		Key   string `json:"key,omitempty"`
		Value string `json:"value,omitempty"`
	} `json:"user_defined_data,omitempty"`
	ValidationType      string `json:"validation_type,omitempty"`
	ReferenceMask       string `json:"reference_mask,omitempty"`
	AcceptanceQualifier string `json:"acceptance_qualifier,omitempty"`
}

type Data struct {
	Attributes     Attributes `json:"attributes,omitempty"`
	CreatedOn      time.Time  `json:"created_on,omitempty"`
	ID             string     `json:"id,omitempty"`
	ModifiedOn     time.Time  `json:"modified_on,omitempty"`
	OrganisationID string     `json:"organisation_id,omitempty"`
	Type           string     `json:"type,omitempty"`
	Version        float64    `json:"version,omitempty"`
}

type Links struct {
	First string `json:"first,omitempty"`
	Last  string `json:"last,omitempty"`
	Self  string `json:"self,omitempty"`
}

type CustomException struct {
	ErrorMessage string `json:"error_message"`
}

//endregion

//region FETCH MODELS

type GetAccountByIdBackendResult struct {
	Data  Data `json:"data"`
	Links `json:"links"`
}

type GetAccountByIdResult struct {
	CreatedOn  time.Time  `json:"created_on"`
	Attributes Attributes `json:"attributes"`
}

//endregion

//region CREATE MODELS

type CreateAccountBackendResult struct {
	Data  Data `json:"data"`
	Links `json:"links"`
}

type CreateAccountResult struct {
	AccountId  string     `json:"account_id"`
	CreatedOn  time.Time  `json:"created_on"`
	Attributes Attributes `json:"attributes"`
}

type CreateAccountRequest struct {
	Attributes     Attributes `json:"attributes"`
	OrganisationID string     `json:"organisation_id"`
}

type CreateAccountBackendRequest struct {
	Data Data `json:"data"`
}

//endregion

//region DELETE MODELS

type DeleteAccountResult struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

//endregion
