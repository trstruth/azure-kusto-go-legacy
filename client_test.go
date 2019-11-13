package kusto

import (
	"fmt"
	"testing"
)

func TestGenerateNewQueryRequest(t *testing.T) {
	clusterName := "testCluster"
	clientID := "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	secret := "xxxxx"
	tenantID := "microsoft.com"

	conn, err := NewConnectionWithAADApplicationKeyAuth(
		clusterName,
		clientID,
		secret,
		tenantID,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	query := "TestTable | take 10"
	database := "testDB"

	c := NewClient(conn)
	req, err := c.generateNewQueryRequest(query, database)
	if  err != nil {
		t.Fatalf(err.Error())
	}

	if req.Method != "POST" {
		t.Fatalf("request method was incorrect: want POST got %s", req.Method)
	}

	expectedConnectionURL := fmt.Sprintf("%s/v1/rest/query", c.connection.url)
	if req.URL.String() != expectedConnectionURL {
		t.Fatalf("request URL was incorrect: want %s, got %s", expectedConnectionURL, req.URL.String())
	}

	expectedHeaders := make(map[string][]string)
	expectedHeaders["Accept"] = []string{"application/json"}
	expectedHeaders["Authorization"] = []string{fmt.Sprintf("Bearer %s", c.connection.servicePrincipalToken.Token().AccessToken)}
	expectedHeaders["Host"] = []string{c.connection.url.Hostname()}
	expectedHeaders["Content-Type"] = []string{"application/json; charset=utf-8"}

	for headerKey, expectedHeaderVal := range expectedHeaders {
		actualHeaderVal := req.Header[headerKey]
		if !stringSlicesAreEqual(actualHeaderVal, expectedHeaderVal) {
			t.Fatalf("the key %s in the generated request was incorrect: want %s, got %s", headerKey, expectedHeaderVal, actualHeaderVal)
		}
	}
}

func TestGenerateNewIngestRequest(t *testing.T) {
	clusterName := "testCluster"
	clientID := "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	secret := "xxxxx"
	tenantID := "microsoft.com"

	conn, err := NewConnectionWithAADApplicationKeyAuth(
		clusterName,
		clientID,
		secret,
		tenantID,
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	data := `{"Col": "Tristan","Col2": "Test","Col3": 123,"Col4": "2019-10-14 03:10:07.1932960"}`
	database := "TestDB"
	table := "TestTable"
	mappingName := "Mapping1"


	c := NewClient(conn)
	req, err := c.generateNewIngestRequest(data, database, table, mappingName)
	if  err != nil {
		t.Fatalf(err.Error())
	}

	if req.Method != "POST" {
		t.Fatalf("request method was incorrect: want POST got %s", req.Method)
	}

	expectedConnectionURL := fmt.Sprintf("%s/v1/rest/ingest/%s/%s?streamFormat=Json&mappingName=%s", c.connection.url, database, table, mappingName)
	if req.URL.String() != expectedConnectionURL {
		t.Fatalf("request URL was incorrect: want %s, got %s", expectedConnectionURL, req.URL.String())
	}

	expectedHeaders := make(map[string][]string)
	expectedHeaders["Authorization"] = []string{fmt.Sprintf("Bearer %s", c.connection.servicePrincipalToken.Token().AccessToken)}
	expectedHeaders["Host"] = []string{c.connection.url.Hostname()}

	for headerKey, expectedHeaderVal := range expectedHeaders {
		actualHeaderVal := req.Header[headerKey]
		if !stringSlicesAreEqual(actualHeaderVal, expectedHeaderVal) {
			t.Fatalf("the key %s in the generated request was incorrect: want %s, got %s", headerKey, expectedHeaderVal, actualHeaderVal)
		}
	}
}

func stringSlicesAreEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}