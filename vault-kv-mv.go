package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	vault "github.com/hashicorp/vault/api"
)

type vaultClient struct {
	logical *vault.Logical
}

// OldNewPaths : Returns a map that has the old path as the key with a value of the new path
func OldNewPaths(leafs []string, source string, destination string) (paths map[string]string) {
	paths = map[string]string{}

	for _, v := range leafs {
		if strings.HasSuffix(source, "/") && !strings.HasSuffix(destination, "/") {
			// Move all secrets under dir secret/old/ under dir secret/new
			d := fmt.Sprintf("%s/", destination)
			paths[v] = strings.Replace(v, source, d, 1)
		} else if strings.HasSuffix(source, "/") && strings.HasSuffix(destination, "/") {
			// Move all secrets under dir secret/old/ to dir secret/new/
			paths[v] = strings.Replace(v, source, destination, 1)
		} else if !strings.HasSuffix(source, "/") && !strings.HasSuffix(destination, "/") {
			// Rename secret secret/old to secret/new
			paths[v] = destination
		} else if !strings.HasSuffix(source, "/") && strings.HasSuffix(destination, "/") {
			// Move secret secret/old under to secret/new/old
			paths[v] = fmt.Sprintf("%s%s", destination, path.Base(source))
		} else {
			// Should never get here
			log.Fatalf("Not sure what to do")
		}
	}
	return paths
}

// AppendDirLeafs : Recursively find leafs if the source is a dir
func (vc *vaultClient) AppendDirLeafs(secrets vault.Secret, source string) (leafs []string) {
	keys := secrets.Data["keys"].([]interface{})
	for _, v := range keys {
		if strings.HasSuffix(v.(string), "/") {
			leafs = append(leafs, vc.FindLeafs(fmt.Sprintf("%v%v", source, v))...)
		} else {
			leafs = append(leafs, fmt.Sprintf("%v%v", source, v))
		}
	}
	return leafs
}

// FindLeafs : Find all keys using the source path supplied by the operator as the starting point
func (vc *vaultClient) FindLeafs(source string) (leafs []string) {
	if strings.HasSuffix(source, "/") {
		listSecret, err := vc.logical.List(source)
		if err != nil || listSecret == nil {
			log.Fatalf("Failed to list %v. Does it exist?", source)
		}
		leafs = vc.AppendDirLeafs(*listSecret, source)
	} else {
		leafs = append(leafs, source)
	}
	return leafs
}

// Move : Creates new entries and then deletes the older ones
func (vc *vaultClient) Move(keys map[string]string) {
	for oldPath, newPath := range keys {
		secret, err := vc.logical.Read(oldPath)
		if err != nil || secret == nil {
			log.Fatalf("Could not read secret %v. Does it exist?", oldPath)
		}

		log.Printf("Writing to new path %v\n", newPath)
		_, err = vc.logical.Write(newPath, secret.Data)
		if err != nil {
			log.Fatalf("Failed to write %v. Try again after fixing the problem.", newPath)
		}

		log.Printf("Deleting old path %v\n", oldPath)
		_, err = vc.logical.Delete(oldPath)
		if err != nil {
			log.Fatalf("Failed to delete old key%v. You will need to manually delete this key after fixing the problem.", oldPath)
		}
	}
}

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), `vault-kv-mv is a tool for moving or renaming secrets in HashiCorp Vault's KV secrets engine.

Usage:
  %s <source_path> <destination_path>

Arguments:
  source_path
      The current path of the secret or directory. Use a trailing '/' to indicate a directory.
  destination_path
      The new path for the secret or directory. Use a trailing '/' to indicate a directory.

Examples:
  # Rename a secret
  %s secret/foo secret/bar

  # Move a secret into a new directory
  %s secret/foo secret/new/

  # Move all secrets from one directory to another
  %s secret/old/ secret/new/

Authentication:
  The tool uses the standard Vault environment variables (e.g., VAULT_ADDR, VAULT_TOKEN).
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
	}

	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		flag.Usage()
		log.Fatal("\n\nInvalid number of arguments. Please specify source and destination paths.")
	}
	source := args[0]
	destination := args[1]

	if source == destination {
		log.Fatalf("source (%s) and destination (%s) are identical. Nothing to do", source, destination)
	}

	config := vault.DefaultConfig()
	config.Timeout = time.Second * 5

	client, err := vault.NewClient(config)
	if err != nil {
		fmt.Printf("Failed to create vault client: %s\n", err)
	}
	vc := vaultClient{
		logical: client.Logical(),
	}

	leafs := vc.FindLeafs(source)
	vc.Move(OldNewPaths(leafs, source, destination))
}
