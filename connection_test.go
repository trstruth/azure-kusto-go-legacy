package kusto

import (
	"testing"
)

func TestNewConnectionWithAADApplicationKeyAuth(t *testing.T) {
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

	if clusterName != conn.clusterName {
		t.Fatalf("clusterName didn't match, want %s got %s", clusterName, conn.clusterName)
	}

	if clientID != conn.clientID {
		t.Fatalf("clusterName didn't match, want %s got %s", clientID, conn.clientID)
	}

	if secret != conn.secret {
		t.Fatalf("aadAppSecret didn't match, want %s got %s", secret, conn.secret)
	}

	if tenantID != conn.tenantID {
		t.Fatalf("authority didn't match, want %s got %s", tenantID, conn.tenantID)
	}
}
