package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
)

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

const URL = "http://localhost:8080/v1/organisation/accounts/"

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

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	switch r.Method {
	case http.MethodGet:
		Fetch(w, r)
		return
	case http.MethodPut:
		Create(w, r)
		return
	case http.MethodDelete:
		Delete(w, r)
		return
	default:
		http.NotFound(w, r)
		return
	}
}

func Fetch(w http.ResponseWriter, r *http.Request) {
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	accountId := r.URL.Query().Get("account_id")

	url := URL + accountId
	response, err := c.Get(url)

	if err != nil {
		http.Error(w, err.Error(), response.StatusCode)
		return
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		var out bytes.Buffer
		json.Indent(&out, body, "", "  ")

		w.WriteHeader(response.StatusCode)
		w.Write(out.Bytes())
		return
	}

	var backendResult GetAccountByIdBackendResult
	errUnmarshal := json.Unmarshal(body, &backendResult)

	if errUnmarshal != nil {
		http.Error(w, errUnmarshal.Error(), http.StatusInternalServerError)
		return
	}

	//map
	var result GetAccountByIdResult
	result.Attributes = backendResult.Data.Attributes
	result.CreatedOn = backendResult.Data.CreatedOn

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(response.StatusCode)
	w.Write(jsonBytes)
}

func Create(w http.ResponseWriter, r *http.Request) {
	c := http.Client{Timeout: time.Duration(1) * time.Second}

	requestBody := &CreateAccountRequest{}
	err := json.NewDecoder(r.Body).Decode(requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestBackend := &CreateAccountBackendRequest{}
	requestBackend.Data.ID = uuid.NewString()
	requestBackend.Data.Type = "accounts"
	requestBackend.Data.OrganisationID = requestBody.OrganisationID
	requestBackend.Data.Attributes = requestBody.Attributes

	accountJson, err := json.Marshal(requestBackend)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := URL

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(accountJson))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := c.Do(req)

	if err != nil {
		http.Error(w, err.Error(), response.StatusCode)
		return
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		var out bytes.Buffer
		json.Indent(&out, body, "", "  ")

		w.WriteHeader(response.StatusCode)
		w.Write(out.Bytes())
		return
	}

	var backendResult CreateAccountBackendResult
	errUnmarshal := json.Unmarshal(body, &backendResult)

	if errUnmarshal != nil {
		http.Error(w, errUnmarshal.Error(), http.StatusInternalServerError)
		return
	}

	var result CreateAccountResult
	result.AccountId = backendResult.Data.ID
	result.Attributes = backendResult.Data.Attributes
	result.CreatedOn = backendResult.Data.CreatedOn

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(response.StatusCode)
	w.Write(jsonBytes)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	accountId := r.URL.Query().Get("account_id")
	version := r.URL.Query().Get("version")

	if len(version) <= 0 {
		version = "0"
	}

	url := URL + accountId + "?version=" + version
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response, err := c.Do(req)

	if err != nil {
		http.Error(w, err.Error(), response.StatusCode)
		return
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		var out bytes.Buffer
		json.Indent(&out, body, "", "  ")

		w.WriteHeader(response.StatusCode)
		w.Write(out.Bytes())
		return
	}
	var result DeleteAccountResult
	result.Message = "Account ID " + accountId + " removed with success"
	result.Success = statusOK

	jsonResult, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResult)
	return

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/accounts", ServeHTTP)
	http.ListenAndServe("localhost:8081", mux)
}
