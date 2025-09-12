package main

import (
	"log"

	approvalserver "antware.xyz/jitaccess/internal/approvalserver"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func main() {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("error creating in-cluster config: %v", err)
	}

	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("error creating dynamic client: %v", err)
	}

	srv := approvalserver.NewServer(client)
	if err := srv.Start(":8080"); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
