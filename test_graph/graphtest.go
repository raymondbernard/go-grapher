// The package creates a sample graph of two nodes
package main

import (
	"os"
	"log"
	"net/http"
	"github.com/raymondbernard/go-grapher/gographer"
	
)


func main() {
	// make sure you set you export your gopath! 
	gopath := os.Getenv("GOPATH")

	graph := gographer.NewG()
	rootServeDir := gopath + "/src/github.com/raymondbernard/go-grapher@v0.1.0/root_serve_dir/"

	// (ID, NodeStringID, GroupName, Size)
	graph.AddNode(1, "NodeStringID", 100, 1)
	graph.AddNode(2, "NodeStringID", 100, 1)

	// (Source, Target, EdgeID, weight )
	graph.AddEdge(1, 2, 0, 1)
	graph.AddEdge(2, 1, 100, 15)

	graph.DumpJSON("graph.json")
	log.Println("Graph created, go visit at localhost:8080")

	panic(http.ListenAndServe(":8080", http.FileServer(http.Dir(rootServeDir))))
}
