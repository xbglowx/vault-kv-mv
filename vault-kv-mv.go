package main

import (
	"fmt"
	"log"

	"flag"
	"github.com/hashicorp/vault/api"
	"strings"
)

func OldNewPaths(leafs []string, source string, destination string) (paths map[string]string) {
	paths = map[string]string{}
	for _, v := range leafs {
		paths[v] = strings.Replace(v, source, destination, 1)
	}
	return paths
}

func FindLeafs(logical api.Logical, source string) (leafs []string) {
	if s, _ := logical.List(source); s != nil {
		secret, err := logical.List(source)

		if err != nil {
			log.Fatalf("Failed to list on %v: %v", source, err)
		}

		keys := secret.Data["keys"].([]interface{})
		for _, v := range keys {
			if strings.HasSuffix(v.(string), "/") {
				leafs = append(leafs, FindLeafs(logical, fmt.Sprintf("%v/%v", source, strings.TrimSuffix(v.(string), "/")))...)
			} else {
				leafs = append(leafs, fmt.Sprintf("%v/%v", source, v))
			}
		}
	} else if s, _ := logical.Read(source); s != nil {
		leafs = append(leafs, source)
	} else {
		log.Fatalf("Source %v is not a valid key in vault", source)
	}
	return leafs
}

func Move(logical api.Logical, keys map[string]string) {
	for oldPath, newPath := range keys {
		secret, err := logical.Read(oldPath)
		if err != nil || secret == nil {
			log.Fatalf("Could not read secret %v. Try again after fixing the problem: %v", oldPath, err)
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

	leafs := FindLeafs(*logical, source)
	Move(*logical, OldNewPaths(leafs, source, destination))
}
