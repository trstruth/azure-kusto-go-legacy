package kusto

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

const (
	clusterName = "testCluster"
	clientID = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	secret = "xxxxx"
	tenantID = "microsoft.com"
	query = "TestTable | take 10"
	database = "testDB"
	data = `{"Col": "Test","Col2": "Test2","Col3": 123,"Col4": "2019-10-14 03:10:07.1932960"}`
	table = "TestTable"
	mappingName = "Mapping1"
)

func TestGenerateNewQueryRequest(t *testing.T) {
	conn, err := NewConnectionWithAADApplicationKeyAuth(
		clusterName,
		clientID,
		secret,
		tenantID,
	)
	if err != nil {
		t.Errorf(err.Error())
	}

	c := NewClient(conn)
	req, err := c.generateNewQueryRequest(query, database)
	if err != nil {
		t.Errorf(err.Error())
	}

	if req.Method != "POST" {
		t.Errorf("request method was incorrect: want POST got %s", req.Method)
	}

	escapedQueryString := strings.ReplaceAll(query, `"`, `\"`)
	expectedData := fmt.Sprintf(`{"csl": "%s", "db": "%s"}`, escapedQueryString, database)
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	actualData := buf.String()

	if expectedData != actualData {
		t.Errorf("Request body data didn't match: want %s, got %s", expectedData, actualData)
	}

	expectedConnectionURL := fmt.Sprintf("%s/v1/rest/query", c.connection.url)
	if req.URL.String() != expectedConnectionURL {
		t.Errorf("request URL was incorrect: want %s, got %s", expectedConnectionURL, req.URL.String())
	}

	expectedHeaders := make(map[string][]string)
	expectedHeaders["Accept"] = []string{"application/json"}
	expectedHeaders["Authorization"] = []string{fmt.Sprintf("Bearer %s", c.connection.servicePrincipalToken.Token().AccessToken)}
	expectedHeaders["Host"] = []string{c.connection.url.Hostname()}
	expectedHeaders["Content-Type"] = []string{"application/json; charset=utf-8"}

	for headerKey, expectedHeaderVal := range expectedHeaders {
		actualHeaderVal := req.Header[headerKey]
		if !reflect.DeepEqual(actualHeaderVal, expectedHeaderVal) {
			t.Errorf("the key %s in the generated request was incorrect: want %s, got %s", headerKey, expectedHeaderVal, actualHeaderVal)
		}
	}
}

func TestGenerateNewIngestRequest(t *testing.T) {
	conn, err := NewConnectionWithAADApplicationKeyAuth(
		clusterName,
		clientID,
		secret,
		tenantID,
	)
	if err != nil {
		t.Errorf(err.Error())
	}

	c := NewClient(conn)
	req, err := c.generateNewIngestRequest(data, database, table, mappingName)
	if err != nil {
		t.Errorf(err.Error())
	}

	if req.Method != "POST" {
		t.Errorf("request method was incorrect: want POST got %s", req.Method)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	actualData := buf.String()

	if data != actualData {
		t.Errorf("Request body data didn't match: want %s, got %s", data, actualData)
	}

	expectedConnectionURL := fmt.Sprintf("%s/v1/rest/ingest/%s/%s?streamFormat=Json&mappingName=%s", c.connection.url, database, table, mappingName)
	if req.URL.String() != expectedConnectionURL {
		t.Errorf("request URL was incorrect: want %s, got %s", expectedConnectionURL, req.URL.String())
	}

	expectedHeaders := make(map[string][]string)
	expectedHeaders["Authorization"] = []string{fmt.Sprintf("Bearer %s", c.connection.servicePrincipalToken.Token().AccessToken)}
	expectedHeaders["Host"] = []string{c.connection.url.Hostname()}

	for headerKey, expectedHeaderVal := range expectedHeaders {
		actualHeaderVal := req.Header[headerKey]
		if !reflect.DeepEqual(actualHeaderVal, expectedHeaderVal) {
			t.Errorf("the key %s in the generated request was incorrect: want %s, got %s", headerKey, expectedHeaderVal, actualHeaderVal)
		}
	}
}

func TestExecuteQuery(t *testing.T) {
	conn, err := NewConnectionWithAADApplicationKeyAuth(
		clusterName,
		clientID,
		secret,
		tenantID,
	)
	if err != nil {
		t.Errorf(err.Error())
	}

	KustoAPIRequestMock := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		fileBytes, err := ioutil.ReadFile("testdata/ExecuteQueryResponseBody.json")
		if err != nil {
			return nil, err
		}

		mockResp := &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBuffer(fileBytes)),
		}
		return mockResp, nil
	})
	mockHTTPClient := &http.Client{
		Transport: KustoAPIRequestMock,
	}

	c := NewClient(conn, SetHTTPClient(mockHTTPClient))

	// disable spt auto refresh to prevent the spt from making a real http request
	// to the aad endpoint with our test clientID/secret
	c.connection.servicePrincipalToken.SetAutoRefresh(false)

	result, err := c.ExecuteQuery(query, database)
	if err != nil {
		t.Errorf(err.Error())
	}

	resultTable := result.Tables[0]

	expectedColumns := []Column{
		Column{"Col", "String"},
		Column{"Col2", "String"},
		Column{"Col3", "Int32"},
		Column{"Col4", "DateTime"},
	}

	expectedRows := []Row{
		Row{nil, nil, nil, nil},
		// note the use of float here - it seems that despite Kustos best attempt to annotate
		// the types via the columns object in the response, go's JSON unmarshaller defaults
		// to unmarshalling numbers to go's float64
		Row{"test", "2k", float64(123), "2019-10-14T03:10:07.193296Z"},
	}

	if !reflect.DeepEqual(expectedColumns, resultTable.Columns) {
		t.Errorf("Columns of result table did not match: want %s, got %s", expectedColumns, resultTable.Columns)
	}

	if !reflect.DeepEqual(expectedRows, resultTable.Rows) {
		t.Errorf("Rows of result table did not match: want %s, got %s", expectedRows, resultTable.Rows)
	}
}

func TestIngestData(t *testing.T) {
	conn, err := NewConnectionWithAADApplicationKeyAuth(
		clusterName,
		clientID,
		secret,
		tenantID,
	)
	if err != nil {
		t.Errorf(err.Error())
	}

	KustoAPIRequestMock := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		fileBytes, err := ioutil.ReadFile("testdata/IngestDataResponseBody.json")
		if err != nil {
			return nil, err
		}

		mockResp := &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBuffer(fileBytes)),
		}
		return mockResp, nil
	})
	mockHTTPClient := &http.Client{
		Transport: KustoAPIRequestMock,
	}

	c := NewClient(conn, SetHTTPClient(mockHTTPClient))

	// disable spt auto refresh to prevent the spt from making a real http request
	// to the aad endpoint with our test clientID/secret
	c.connection.servicePrincipalToken.SetAutoRefresh(false)

	data := `{"Col": "Tristan","Col2": "Test","Col3": 123,"Col4": "2019-10-14 03:10:07.1932960"}`
	database := ""
	table := "TestTable"
	mappingName := "Mapping1"

	result, err := c.IngestData(data, database, table, mappingName)
	if err != nil {
		t.Errorf(err.Error())
	}

	resultTable := result.Tables[0]

	expectedColumns := []Column{
		Column{"ConsumedRecordsCount", "Int64"},
		Column{"UpdatePolicyStatus", "String"},
		Column{"UpdatePolicyFailureCode", "String"},
		Column{"UpdatePolicyFailureReason", "String"},
	}

	expectedRows := []Row{
		// note the use of float here - it seems that despite Kustos best attempt to annotate
		// the types via the columns object in the response, go's JSON unmarshaller defaults
		// to unmarshalling numbers to go's float64
		Row{float64(0), "Inactive", "Unknown", nil},
	}

	if !reflect.DeepEqual(expectedColumns, resultTable.Columns) {
		t.Errorf("Columns of result table did not match: want %s, got %s", expectedColumns, resultTable.Columns)
	}

	if !reflect.DeepEqual(expectedRows, resultTable.Rows) {
		t.Errorf("Rows of result table did not match: want %s, got %s", expectedRows, resultTable.Rows)
	}
}

// RoundTripFunc is the function which will mock the actual req/resp
// round trip
type RoundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip is a method of the function type RoundTripFunc, such that
// it satisfies the RoundTripper interface
func (fn RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
