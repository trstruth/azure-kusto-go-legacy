package kusto

import (
	"fmt"
	"net/url"

	"github.com/Azure/go-autorest/autorest/adal"
)

// Connection is a struct which describes a kusto connection
type Connection struct {
	clusterName           string
	clientID              string
	secret                string
	tenantID              string
	url                   *url.URL
	servicePrincipalToken *adal.ServicePrincipalToken
}

// NewConnectionWithAADApplicationKeyAuth constructs a new Connection struct using AAD Application ID and Key
func NewConnectionWithAADApplicationKeyAuth(clusterName, clientID, secret, tenantID string) (*Connection, error) {
	c := Connection{}

	c.clusterName = clusterName
	c.clientID = clientID
	c.secret = secret
	c.tenantID = tenantID

	rawURLString := fmt.Sprintf("https://%s.kusto.windows.net", clusterName)
	parsedURL, err := url.Parse(rawURLString)
	if err != nil {
		return nil, err
	}
	c.url = parsedURL

	oauthConfig, err := adal.NewOAuthConfig("https://login.microsoftonline.com", tenantID)
	if err != nil {
		return nil, err
	}
	servicePrincipalToken, err := adal.NewServicePrincipalToken(
		*oauthConfig,
		clientID,
		secret,
		fmt.Sprintf("%s://%s", c.url.Scheme, c.url.Host),
	)
	if err != nil {
		return nil, err
	}
	c.servicePrincipalToken = servicePrincipalToken

	return &c, nil
}
