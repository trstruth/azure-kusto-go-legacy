# Microsoft Azure Kusto  (Azure Data Explorer) SDK  for Go

### Install
To install via the Python Package Index (PyPI), type:
* `go get github.com/trstruth/azure-kusto-go`

### Minimum Requirements
* go 1.12.9
* See go.mod for dependencies

### Authentication methods:
* AAD application - Provide app ID and app secret to Kusto client.

### Usage:
```go
package main

import (
    kusto "github.com/trstruth/azure-kusto-go"
)

func main() {
    clusterName := "MyCluster"
    clientID := "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    secret := "xxxxx" // just an example, don't check secrets into source control
    tenantID := "microsoft.com"

    conn := kusto.NewConnectionWithAADApplicationKeyAuth(
        clusterName,
        clientID,
        secret,
        tenantID,
    )
    
    c := kusto.NewClient(conn)
    
    query := "MyTable | take 10"
    database := "MyDatabase"
    result, err := c.executeQuery(query, database)
    if err != nil {
        fmt.Errorf(err.Error())
    }
    
    resultTable := result.Tables[0]
    fmt.Println(resultTable.Columns)
    fmt.Println(resultTable.Rows)
}
```

## Looking for SDKs for other languages/platforms?
- [Node](https://github.com/azure/azure-kusto-node)
- [Java](https://github.com/azure/azure-kusto-java)
- [.NET](https://docs.microsoft.com/en-us/azure/kusto/api/netfx/about-the-sdk)
- [Python](https://github.com/Azure/azure-kusto-python)