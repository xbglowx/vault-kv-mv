package main

import (
	"context"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testVaultServer creates a test vault container and returns a configured API
// client and closer function.
func testVaultServer(t *testing.T) (*api.Client, func()) {
	t.Helper()

	client, _, closer := testVaultServerUnseal(t)
	return client, closer
}

// testVaultServerUnseal creates a test vault container and returns a configured
// API client, list of unseal keys (as strings), and a closer function.
func testVaultServerUnseal(t *testing.T) (*api.Client, []string, func()) {
	t.Helper()

	ctx := context.Background()

	// Start Vault container
	req := testcontainers.ContainerRequest{
		Image:        "hashicorp/vault:1.19.5",
		ExposedPorts: []string{"8200/tcp"},
		Env: map[string]string{
			"VAULT_DEV_ROOT_TOKEN_ID": "root",
			"VAULT_DEV_LISTEN_ADDRESS": "0.0.0.0:8200",
		},
		Cmd: []string{"vault", "server", "-dev"},
		WaitingFor: wait.ForLog("Development mode should NOT be used in production installations!").
			WithStartupTimeout(30 * time.Second),
	}

	vaultContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start vault container: %v", err)
	}

	// Get container port
	mappedPort, err := vaultContainer.MappedPort(ctx, "8200")
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	// Get container host
	host, err := vaultContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	// Configure Vault client
	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	
	client, err := api.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create vault client: %v", err)
	}

	// Set token
	client.SetToken("root")

	// Wait for Vault to be ready and enable KV backends
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		_, err := client.Sys().Health()
		if err == nil {
			break
		}
		if i == maxRetries-1 {
			t.Fatalf("Vault not ready after %d retries: %v", maxRetries, err)
		}
		time.Sleep(1 * time.Second)
	}

	// Disable the default KV v2 mount at secret/ and enable KV v1
	if err := client.Sys().Unmount("secret"); err != nil {
		t.Logf("Warning: could not unmount default secret backend: %v", err)
	}

	// Enable KV v1 backend at secret/
	if err := client.Sys().Mount("secret", &api.MountInput{
		Type: "kv",
		Options: map[string]string{
			"version": "1",
		},
	}); err != nil {
		t.Fatalf("Failed to mount kv v1 backend: %v", err)
	}

	// For dev mode, we don't have actual unseal keys, so we return empty keys
	unsealKeys := []string{}

	closer := func() {
		if err := vaultContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate vault container: %v", err)
		}
	}

	return client, unsealKeys, closer
}



func TestRenameSecret(t *testing.T) {
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
	_, err := logical.Write(source, data)
	if err != nil {
		t.Errorf("Failed to write/seed data during testing. %v", err)
	}

	vc := vaultClient{
		logical: logical,
	}

	leafs := vc.FindLeafs(source)
	vc.Move(OldNewPaths(leafs, source, destination))

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

func TestMoveDirToDirTrailingSlash(t *testing.T) {
	client, closer := testVaultServer(t)
	defer closer()

	source := "secret/old/"
	destination := "secret/new/"

	secretName := "foo/bar"
	oldSecret := fmt.Sprintf("%v%v", source, secretName)
	newSecret := fmt.Sprintf("%v%v", destination, secretName)

	key := "test"
	value := "test"
	data := map[string]interface{}{
		key: value,
	}

	logical := client.Logical()
	_, err := logical.Write(oldSecret, data)
	if err != nil {
		t.Errorf("Failed to write/seed data during testing. %v", err)
	}

	vc := vaultClient{
		logical: client.Logical(),
	}

	leafs := vc.FindLeafs(source)
	vc.Move(OldNewPaths(leafs, source, destination))

	secret, err := logical.Read(newSecret)
	if err != nil {
		t.Errorf("Error while reading vault key %v: %v", newSecret, err)
	}

	if secret.Data[key] != value {
		t.Errorf("Expected key/value of %v:%v for %v. Got %v instead", key, value, newSecret, secret.Data[newSecret])
	}

	secret, err = logical.Read(oldSecret)
	if err != nil {
		t.Errorf("Failed while checking vault for the old path: %v", err)
	}

	if secret != nil {
		t.Errorf("Expected path %v to be deleted. But, it still exists.", oldSecret)
	}
}

func TestMoveDirToDirNoTrailingSlash(t *testing.T) {
	client, closer := testVaultServer(t)
	defer closer()

	source := "secret/old/"
	destination := "secret/new"

	secretName := "foo/bar"
	oldSecret := fmt.Sprintf("%v%v", source, secretName)
	newSecret := fmt.Sprintf("%v/%v", destination, secretName)

	key := "test"
	value := "test"
	data := map[string]interface{}{
		key: value,
	}

	logical := client.Logical()
	_, err := logical.Write(oldSecret, data)
	if err != nil {
		t.Errorf("Failed to write/seed data during testing. %v", err)
	}

	vc := vaultClient{
		logical: client.Logical(),
	}

	leafs := vc.FindLeafs(source)
	vc.Move(OldNewPaths(leafs, source, destination))

	secret, err := logical.Read(newSecret)
	if err != nil {
		t.Errorf("Error while reading vault key %v: %v", newSecret, err)
	}

	if secret.Data[key] != value {
		t.Errorf("Expected key/value of %v:%v for %v. Got %v instead", key, value, newSecret, secret.Data[newSecret])
	}

	secret, err = logical.Read(oldSecret)
	if err != nil {
		t.Errorf("Failed while checking vault for the old path: %v", err)
	}

	if secret != nil {
		t.Errorf("Expected path %v to be deleted. But, it still exists.", oldSecret)
	}
}

func TestMoveSecretToDir(t *testing.T) {
	client, closer := testVaultServer(t)
	defer closer()

	source := "secret/foo"
	destination := "secret/bar/"
	newDestinationFile := fmt.Sprintf("%s%s", destination, path.Base(source))

	key := "test"
	value := "test"
	data := map[string]interface{}{
		key: value,
	}

	logical := client.Logical()
	_, err := logical.Write(source, data)
	if err != nil {
		t.Errorf("Failed to write/seed data during testing. %v", err)
	}

	vc := vaultClient{
		logical: client.Logical(),
	}

	leafs := vc.FindLeafs(source)
	vc.Move(OldNewPaths(leafs, source, destination))

	secret, err := logical.Read(newDestinationFile)
	if err != nil {
		t.Errorf("Error while reading vault key %v: %v", newDestinationFile, err)
	}

	if secret.Data[key] != value {
		t.Errorf("Expected key/value of %v:%v for %v. Got %v instead", key, value, newDestinationFile, secret.Data[newDestinationFile])
	}

	secret, err = logical.Read(source)
	if err != nil {
		t.Errorf("Failed while checking vault for the old path: %v", err)
	}

	if secret != nil {
		t.Errorf("Expected path %v to be deleted. But, it still exists.", source)
	}
}
