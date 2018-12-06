package main

import (
	"encoding/base64"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/audit"
	"github.com/hashicorp/vault/builtin/logical/database"
	"github.com/hashicorp/vault/builtin/logical/pki"
	"github.com/hashicorp/vault/builtin/logical/transit"
	"github.com/hashicorp/vault/helper/builtinplugins"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/vault"

	"fmt"
	auditFile "github.com/hashicorp/vault/builtin/audit/file"
	credUserpass "github.com/hashicorp/vault/builtin/credential/userpass"
	vaulthttp "github.com/hashicorp/vault/http"
)

// testVaultServer creates a test vault cluster and returns a configured API
// client and closer function.
func testVaultServer(t testing.TB) (*api.Client, func()) {
	t.Helper()

	client, _, closer := testVaultServerUnseal(t)
	return client, closer
}

// testVaultServerUnseal creates a test vault cluster and returns a configured
// API client, list of unseal keys (as strings), and a closer function.
func testVaultServerUnseal(t testing.TB) (*api.Client, []string, func()) {
	t.Helper()

	return testVaultServerCoreConfig(t, &vault.CoreConfig{
		DisableMlock: true,
		DisableCache: true,
		CredentialBackends: map[string]logical.Factory{
			"userpass": credUserpass.Factory,
		},
		AuditBackends: map[string]audit.Factory{
			"file": auditFile.Factory,
		},
		LogicalBackends: map[string]logical.Factory{
			"database":       database.Factory,
			"generic-leased": vault.LeasedPassthroughBackendFactory,
			"pki":            pki.Factory,
			"transit":        transit.Factory,
		},
		BuiltinRegistry: builtinplugins.Registry,
	})
}

// testVaultServerCoreConfig creates a new vault cluster with the given core
// configuration. This is a lower-level test helper.
func testVaultServerCoreConfig(t testing.TB, coreConfig *vault.CoreConfig) (*api.Client, []string, func()) {
	t.Helper()

	cluster := vault.NewTestCluster(t, coreConfig, &vault.TestClusterOptions{
		HandlerFunc: vaulthttp.Handler,
	})
	cluster.Start()

	// Make it easy to get access to the active
	core := cluster.Cores[0].Core
	vault.TestWaitActive(t, core)

	// Get the client already setup for us!
	client := cluster.Cores[0].Client
	client.SetToken(cluster.RootToken)

	// Convert the unseal keys to base64 encoded, since these are how the user
	// will get them.
	unsealKeys := make([]string, len(cluster.BarrierKeys))
	for i := range unsealKeys {
		unsealKeys[i] = base64.StdEncoding.EncodeToString(cluster.BarrierKeys[i])
	}

	return client, unsealKeys, func() { defer cluster.Cleanup() }
}

func TestMoveKey(t *testing.T) {
	client, closer := testVaultServer(t)
	defer closer()

	source := "secret/old"
	destination := "secret/new"
	key := "test"
	value := "test"
	data := map[string]interface{}{
		key: value,
	}

	logical := client.Logical()
	logical.Write(source, data)

	leafs := FindLeafs(*logical, source)
	Move(*logical, OldNewPaths(leafs, source, destination))

	secret, err := logical.Read(destination)
	if err != nil {
		t.Errorf("Error while reading vault key %v: %v", source, err)
	}

	if secret.Data[key] != value {
		t.Errorf("Expected key/value of %v:%v for %v. Got %v instead", key, value, source, secret.Data[source])
	}

	secret, err = logical.Read(source)
	if err != nil {
		t.Errorf("Failed while checking vault for the old path: %v", err)
	}

	if secret != nil {
		t.Errorf("Expected path %v to be deleted. But, it still exists.", source)
	}
}

func TestMoveDir(t *testing.T) {
	client, closer := testVaultServer(t)
	defer closer()

	source := "secret/old"
	destination := "secret/new"
	suffix := "foo/bar"
	oldPath := fmt.Sprintf("%v/%v", source, suffix)
	newPath := fmt.Sprintf("%v/%v", destination, suffix)

	key := "test"
	value := "test"
	data := map[string]interface{}{
		key: value,
	}

	logical := client.Logical()
	logical.Write(oldPath, data)

	leafs := FindLeafs(*logical, source)
	Move(*logical, OldNewPaths(leafs, source, destination))

	secret, err := logical.Read(newPath)
	if err != nil {
		t.Errorf("Error while reading vault key %v: %v", newPath, err)
	}

	if secret.Data[key] != value {
		t.Errorf("Expected key/value of %v:%v for %v. Got %v instead", key, value, newPath, secret.Data[newPath])
	}

	secret, err = logical.Read(oldPath)
	if err != nil {
		t.Errorf("Failed while checking vault for the old path: %v", err)
	}

	if secret != nil {
		t.Errorf("Expected path %v to be deleted. But, it still exists.", oldPath)
	}
}
