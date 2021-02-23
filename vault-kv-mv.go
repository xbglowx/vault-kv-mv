package main

import (
	"log"
	"path"

	"flag"
	"fmt"
	"github.com/hashicorp/vault/api"
	"strings"
)

// OldNewPaths : Returns a map that has the old path as the key with a value of the new path
func OldNewPaths(leafs []string, source string, destination string) (paths map[string]string) {
	paths = map[string]string{}

	for _, v := range leafs {
		if strings.HasSuffix(destination, "/") {
			paths[v] = fmt.Sprintf("%s%s", destination, path.Base(source))
		} else {
			paths[v] = strings.Replace(v, source, destination, 1)
		}
	}
	return paths
}

// AppendDirLeafs : Recursively find leafs if the source is a dir
func AppendDirLeafs(secrets api.Secret, logical api.Logical, source string) (leafs []string) {
	keys := secrets.Data["keys"].([]interface{})
	for _, v := range keys {
		if strings.HasSuffix(v.(string), "/") {
			leafs = append(leafs, FindLeafs(logical, fmt.Sprintf("%v/%v", source, strings.TrimSuffix(v.(string), "/")))...)
		} else {
			leafs = append(leafs, fmt.Sprintf("%v/%v", source, v))
		}
	}
	return leafs
}

// FindLeafs : Find all keys using the source path supplied by the operator as the starting point
func FindLeafs(logical api.Logical, source string) (leafs []string) {
	if strings.HasSuffix(source, "/") {
		listSecret, err := logical.List(source)
		if err != nil {
			log.Fatalf("Failed to list %v: %v", source, err)
		}
		leafs = AppendDirLeafs(*listSecret, logical, source)
	} else {
		leafs = append(leafs, source)
	}
	return leafs
}

// Move : Creates new entries and then deletes the older ones
func Move(logical api.Logical, keys map[string]string) {
	for oldPath, newPath := range keys {
		secret, err := logical.Read(oldPath)
		if err != nil || secret == nil {
			log.Fatalf("Could not read secret %v. Does it exist?", oldPath)
		}

		log.Printf("Writing to new path %v\n", newPath)
		_, err = logical.Write(newPath, secret.Data)
		if err != nil {
			log.Fatalf("Failed to write %v. Try again after fixing the problem.", newPath)
		}

		log.Printf("Deleting old path %v\n", oldPath)
		_, err = logical.Delete(oldPath)
		if err != nil {
			log.Fatalf("Failed to delete old key%v. You will need to manually delete this key after fixing the problem.", oldPath)
		}
	}
}

func main() {
	client, err := api.NewClient(nil)
	if err != nil {
		log.Fatalf("Failed to create a vault client: %v", err)
	}
	logical := client.Logical()

	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		log.Fatal("Invalid number of arugments. Need to specify source and destination paths.")
	}
	source := args[0]
	destination := args[1]

	if source == destination {
		log.Fatalf("source (%s) and destination (%s) are identical. Nothing to do", source, destination)
	}

	leafs := FindLeafs(*logical, source)
	Move(*logical, OldNewPaths(leafs, source, destination))
}
