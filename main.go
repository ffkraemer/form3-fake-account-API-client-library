package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/client-library/domain"

	"github.com/google/uuid"
)

const URL = "http://localhost:8080/v1/organisation/accounts"

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

	url := URL + "/" + accountId
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

	var backendResult domain.GetAccountByIdBackendResult
	errUnmarshal := json.Unmarshal(body, &backendResult)

	if errUnmarshal != nil {
		http.Error(w, errUnmarshal.Error(), http.StatusInternalServerError)
		return
	}

	//map
	var result domain.GetAccountByIdResult
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

	requestBody := &domain.CreateAccountRequest{}
	err := json.NewDecoder(r.Body).Decode(requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestBackend := &domain.CreateAccountBackendRequest{}
	requestBackend.Data.ID = uuid.NewString()
	requestBackend.Data.Type = "accounts"
	requestBackend.Data.OrganisationID = requestBody.OrganisationID
	requestBackend.Data.Attributes = requestBody.Attributes

	accountJson, err := json.Marshal(requestBackend)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := http.Post(URL, "application/json", bytes.NewBuffer(accountJson))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	var backendResult domain.CreateAccountBackendResult
	errUnmarshal := json.Unmarshal(body, &backendResult)

	if errUnmarshal != nil {
		http.Error(w, errUnmarshal.Error(), http.StatusInternalServerError)
		return
	}

	var result domain.CreateAccountResult
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

	url := URL + "/" + accountId + "?version=" + version
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
	var result domain.DeleteAccountResult
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
