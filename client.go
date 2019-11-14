package kusto

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Client represents a KustoClient
// connection contains the parameters used to connect to Kusto
// httpClient is the implementation of the httpClient interface used
// to make http requests to Kusto itself
type Client struct {
	connection *Connection
	httpClient *http.Client
}

// Option is a function that modifies an instance of Client.  Pairing this with
// variadic argument syntax in the initializer allows us to configure a new
// instance of Client
type Option func(*Client)

// SetHTTPClient is an option that allows us to configure the http client member
// of a Client instance
func SetHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient constructs a Client struct with a "real" http client in its
// httpClient field.  This function should be called by external users of
// the kusto package
func NewClient(conn *Connection, options ...Option) *Client {
	c := Client{}

	c.connection = conn
	c.httpClient = &http.Client{}

	for _, optionFunc := range options {
		optionFunc(&c)
	}

	return &c
}

// ExecuteQuery runs the supplied query against the configured kusto cluster
// and returns the result data or an error
func (c *Client) ExecuteQuery(query, database string) (*QueryResult, error) {
	err := c.connection.servicePrincipalToken.EnsureFresh()
	if err != nil {
		return nil, err
	}

	req, err := c.generateNewQueryRequest(query, database)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	queryResult := QueryResult{}
	err = json.Unmarshal(bodyBytes, &queryResult)
	if err != nil {
		return nil, err
	}

	return &queryResult, nil
}

// generateNewQueryRequest is a helper used by ExecuteQuery to generate a request
// to be send to Kusto
func (c *Client) generateNewQueryRequest(query, database string) (*http.Request, error) {
	escapedQueryString := strings.ReplaceAll(query, `"`, `\"`)
	jsonStr := fmt.Sprintf(`{"csl": "%s", "db": "%s"}`, escapedQueryString, database)
	jsonBytes := []byte(jsonStr)
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/v1/rest/query", c.connection.url),
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return nil, err
	}

	// required headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.connection.servicePrincipalToken.Token().AccessToken))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", c.connection.url.Hostname())

	return req, nil
}

// IngestData performs streaming ingest of `data` into `database`
func (c *Client) IngestData(data, database, table, mappingName string) (*IngestResult, error) {
	err := c.connection.servicePrincipalToken.EnsureFresh()
	if err != nil {
		return nil, err
	}

	req, err := c.generateNewIngestRequest(data, database, table, mappingName)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	IngestResult := IngestResult{}
	err = json.Unmarshal(bodyBytes, &IngestResult)
	if err != nil {
		return nil, err
	}

	return &IngestResult, nil
}

// generateNewIngestRequest is a helper used by IngestData to generate a request to
// ingest data into kusto
func (c *Client) generateNewIngestRequest(data, database, table, mappingName string) (*http.Request, error) {
	dataBytes := []byte(data)
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/v1/rest/ingest/%s/%s?streamFormat=Json&mappingName=%s", c.connection.url, database, table, mappingName),
		bytes.NewBuffer(dataBytes),
	)
	if err != nil {
		return nil, err
	}

	// required headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.connection.servicePrincipalToken.Token().AccessToken))
	req.Header.Set("Host", c.connection.url.Hostname())

	// optional headers

	return req, nil
}
