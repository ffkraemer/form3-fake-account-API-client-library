package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/client-library/domain"
)

//region CreateTestCase
var organisationIds = []string{
	"0d077184-ca1b-4583-a416-29c9a51cf6e",
	"84385b9c-176d-11ed-861d-0242ac120002"}

var accountIds = []string{
	"3a877792-1783-11ed-861d-0242ac12000",
	"802052e6-182e-11ed-861d-0242ac120002",
	"cc3a78fe-1785-11ed-861d-0242ac120002"}

var createAccountRequest_Client = domain.CreateAccountRequest{
	OrganisationID: organisationIds[1],
	Attributes: domain.Attributes{
		Country:      "GB",
		BaseCurrency: "GBP",
		BankID:       "400300",
		BankIDCode:   "GBDSC",
		Bic:          "NWBKGB22",
		Name: []string{
			"Fábio Fragoso Kraemer Moraes",
		},
		AlternativeNames: []string{
			"Fábio Moraes",
		},
		UserDefinedData: nil,
	},
}

var clientTestCasesCreate = []struct {
	name                   string
	request                domain.CreateAccountRequest
	expected_status_code   int
	expected_message_error string
	generated_ids          []string
}{
	{"ShouldHaveSuccess", createAccountRequest_Client, http.StatusCreated, "", nil},
	{"InvalidOrganisationID", domain.CreateAccountRequest{OrganisationID: organisationIds[0],
		Attributes: createAccountRequest_Client.Attributes}, http.StatusBadRequest, "organisation_id in body must be of type uuid: \"0d077184-ca1b-4583-a416-29c9a51cf6e\"", nil},
	{"CountryIsRequired", domain.CreateAccountRequest{OrganisationID: createAccountRequest_Client.OrganisationID,
		Attributes: domain.Attributes{
			Name: createAccountRequest_Client.Attributes.Name}},
		http.StatusBadRequest, "country in body is required", nil},
	{"NameIsRequired", domain.CreateAccountRequest{OrganisationID: createAccountRequest_Client.OrganisationID,
		Attributes: domain.Attributes{
			Country: createAccountRequest_Client.Attributes.Country}},
		http.StatusBadRequest, "name in body is required", nil},
	{"CountryNotMatches", domain.CreateAccountRequest{OrganisationID: createAccountRequest_Client.OrganisationID,
		Attributes: domain.Attributes{
			Country: "B"}},
		http.StatusBadRequest, "country in body should match", nil},
	{"NameMoreThan140CharsIsInvalid", domain.CreateAccountRequest{OrganisationID: createAccountRequest_Client.OrganisationID,
		Attributes: domain.Attributes{
			Name: []string{MockingMaxLengthString(140)}}},
		http.StatusBadRequest, "in body should be at most 140 chars long", nil}}

var createAccountRequest = domain.CreateAccountBackendRequest{
	Data: domain.Data{
		Type:           "accounts",
		OrganisationID: organisationIds[1],
		ID:             accountIds[1],
		Attributes: domain.Attributes{
			Country: "GB",
			Name: []string{
				"Fábio Fragoso Kraemer Moraes",
			},
		}}}

var testCasesCreate = []struct {
	name                   string
	request                domain.CreateAccountBackendRequest
	expected_status_code   int
	expected_message_error string
}{
	{"ShouldHaveSuccess", createAccountRequest, http.StatusCreated, ""},
	{"ViolatesDuplicateConstraint", createAccountRequest, http.StatusConflict, "Account cannot be created as it violates a duplicate constraint"},
	{"InvalidOrganisationID", domain.CreateAccountBackendRequest{
		Data: domain.Data{
			OrganisationID: organisationIds[0],
			ID:             createAccountRequest.Data.ID,
			Attributes:     createAccountRequest.Data.Attributes}},
		http.StatusBadRequest, "organisation_id in body must be of type uuid: \"0d077184-ca1b-4583-a416-29c9a51cf6e\""},
	{"InvalidAccountID", domain.CreateAccountBackendRequest{
		Data: domain.Data{
			OrganisationID: createAccountRequest.Data.OrganisationID,
			ID:             accountIds[0],
			Attributes:     createAccountRequest_Client.Attributes}}, http.StatusBadRequest, "id in body must be of type uuid: \"3a877792-1783-11ed-861d-0242ac12000\""},
	{"InvalidType", domain.CreateAccountBackendRequest{
		Data: domain.Data{
			OrganisationID: createAccountRequest.Data.OrganisationID,
			ID:             accountIds[2],
			Type:           "acounts",
			Attributes:     createAccountRequest.Data.Attributes}},
		http.StatusBadRequest, "type in body should be one of [accounts]"},
	{"CountryIsRequired", domain.CreateAccountBackendRequest{
		Data: domain.Data{
			OrganisationID: createAccountRequest.Data.OrganisationID,
			ID:             accountIds[2],
			Attributes: domain.Attributes{
				Name: createAccountRequest.Data.Attributes.Name}}},
		http.StatusBadRequest, "country in body is required"},
	{"NameIsRequired", domain.CreateAccountBackendRequest{
		Data: domain.Data{
			OrganisationID: createAccountRequest.Data.OrganisationID,
			ID:             accountIds[2],
			Attributes: domain.Attributes{
				Country: createAccountRequest.Data.Attributes.Country}}},
		http.StatusBadRequest, "name in body is required"},
	{"CountryNotMatches", domain.CreateAccountBackendRequest{
		Data: domain.Data{
			OrganisationID: createAccountRequest.Data.OrganisationID,
			ID:             accountIds[2],
			Attributes: domain.Attributes{
				Name:    createAccountRequest.Data.Attributes.Name,
				Country: "B"}}},
		http.StatusBadRequest, "country in body should match"},
	{"NameMoreThan140CharsIsInvalid", domain.CreateAccountBackendRequest{
		Data: domain.Data{
			OrganisationID: createAccountRequest.Data.OrganisationID,
			ID:             accountIds[2],
			Attributes: domain.Attributes{
				Name: []string{MockingMaxLengthString(140)}}}},
		http.StatusBadRequest, "in body should be at most 140 chars long"}}

//endregion

func TestCreateAccount_ClientLibrary(t *testing.T) {

	for _, tc := range clientTestCasesCreate {
		t.Run("Client_"+tc.name, func(t *testing.T) {
			expected_status_code := tc.expected_status_code

			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(tc.request)
			if err != nil {
				t.Errorf(err.Error())
				return
			}

			r := httptest.NewRequest(http.MethodPut, "/accounts", &buf)
			w := httptest.NewRecorder()
			ServeHTTP(w, r)

			if w.Code != expected_status_code {
				t.Errorf("Expected %d, returned %d", expected_status_code, w.Code)
				return
			}

			//This test is expecting some error
			if len(tc.expected_message_error) > 0 {

				body, err := ioutil.ReadAll(w.Body)
				if err != nil {
					t.Errorf(err.Error())
					return
				}

				var exc domain.CustomException
				errUnmarshal := json.Unmarshal(body, &exc)
				if errUnmarshal != nil {
					t.Errorf(errUnmarshal.Error())
					return
				}

				if !strings.Contains(exc.ErrorMessage, tc.expected_message_error) {
					t.Errorf("Expected %s, returned %s", tc.expected_message_error, exc.ErrorMessage)
					return
				}
			}

			body, err := ioutil.ReadAll(w.Body)
			var result domain.CreateAccountResult
			errUnmarshal := json.Unmarshal(body, &result)
			if errUnmarshal != nil {
				t.Errorf(errUnmarshal.Error())
				return
			}
			accountIds = append(accountIds, result.AccountId)
		})
	}
}

func TestCreateAccount_ServiceAPI(t *testing.T) {

	for _, tc := range testCasesCreate {
		t.Run("ServiceAPI_"+tc.name, func(t *testing.T) {
			expected_status_code := tc.expected_status_code

			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(tc.request)
			if err != nil {
				t.Errorf(err.Error())
				return
			}

			response, err := http.Post(URL, "application/json", &buf)

			if err != nil {
				t.Errorf(err.Error())
				return
			}

			if response.StatusCode != expected_status_code {
				t.Errorf("Expected %d, returned %d", expected_status_code, response.StatusCode)
				t.Errorf("Status Decription: %s", response.Status)
				return
			}

			//This test is expecting some error
			if len(tc.expected_message_error) > 0 {

				body, err := ioutil.ReadAll(response.Body)
				if err != nil {
					t.Errorf(err.Error())
					return
				}

				var exc domain.CustomException
				errUnmarshal := json.Unmarshal(body, &exc)
				if errUnmarshal != nil {
					t.Errorf(errUnmarshal.Error())
					return
				}

				if !strings.Contains(exc.ErrorMessage, tc.expected_message_error) {
					t.Errorf("Expected %s, returned %s", tc.expected_message_error, exc.ErrorMessage)
					return
				}
			}
		})
	}
}

func Test_TearDown(t *testing.T) {

	for _, accountId := range accountIds {
		url := URL + "/" + accountId + "?version=0"
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response, err := c.Do(req)
	}
}

func MockingMaxLengthString(maxLength int) string {
	var str = ""

	for i := 0; i <= maxLength; i++ {
		if len(str) <= maxLength {
			str += strconv.Itoa(i)
		}

		if len(str) > maxLength {
			break
		}
	}

	return str
}
