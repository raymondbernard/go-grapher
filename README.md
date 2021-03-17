# gographer
It is a fork from "github.com/fjukstad/gographer"

IMPORTANT -- This repo is under active developement and should not be used in production until we acheive v1.0.0 
Our initial release is 0.1.0 which is unstable at best. 
https://semver.org/


Fixed various websocket issues. Now we using the standard lib websockets.   We will be using the semantic version system.  Once it hits v 1.0 consider it stable and production ready. Our goal is to use this repo to build out a richer set of visualazions based on d3js and then produce a scalable graph db using go. 

Simple graph package for go. Uses [d3js](https://github.com/mbostock/d3) for visualization and websockets for communication. 


# Run the test visualization

    go run test_graph/visualization.go
    
and visit [localhost:8080](http://localhost:8080) in your browser 

# Using it
Import it:

    import "github.com/raymondbernard/go-grapher/gographer"

Using it:

    graph = gographer.New();
    // (ID, NodeText, GroupID, Size)
    graph.AddNode( 1, "Node Text blah", "1234", 1 )
    graph.AddNode( 2, "Node Text blah 2", "1234", 1 )

    // (Source, Target, EdgeID, weight )
    grap.AddEdge( 1, 2, 0, 1 )
    graph.AddEdge( 2, 1, 100, 15 )

    graph.DumpJSON( "graph.json" )
    http.ListenAndServe( ":8080", http.FileServer( http.Dir( "." ) ) )


Open [localhost:8080](http://localhost:8080) in a webbrowser to view the graph.


Modify the d3js implementation and visualization to your preferences.


# Files

- _gographer.go_ contains source code to get started with a graph library
- _visualizer.go_ contains source code to host the visualizations
- _root_serve_dir/index.html_ contains the empty shell for the visualization
- _root_serve_dir/js/_ contains source code for the visualization
- _root_serve_dir/css/_ contains css files for visualization
- _graph.json_ contains the output from gographer.go (graph).
