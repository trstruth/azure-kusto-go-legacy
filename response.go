package kusto

// QueryResult is the model which contains the fields present in
// the response of an http request to the /v1/rest/query endpoint
// https://kusto.azurewebsites.net/docs/api/rest/response.html#json-encoding-of-a-sequence-of-tables
type QueryResult struct {
	Tables []Table
}

// Table is the model which contains the fields present in
// each element of the Tables slice in QueryResult
type Table struct {
	TableName string
	Columns   []Column
	Rows      []Row
}

// Column is the model which contains the fields present in
// each element of the Columns slice in Table
type Column struct {
	ColumnName string
	DataType   string
	ColumnType string
}

// Row is a slice of empty interfaces, where a given index in that
// slice corresponds to the value at the same index in the Columns
// slice in Table
type Row []interface{}
