// go-grapher creates a graph using golang code. The visuaslization is outputed to html with d3.
// The data from d3 is updated via websockets. No need to refresh the page. 
//  We have updated the websocket to use the std lib and consolidated the methods into a single golang package.  
// by Ray Bernard
// This program is a fork from https://github.com/fjukstad/gographer
package gographer 

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)
// Communication.go Begin
type InitGraphMessage struct {
	Command string `json:"command,string"`
	Graph   string `json:"graph,string"`
}

type AddNodeMessage struct {
	Command string `json:"command,string"`
	Id      int    `json:"id,int"`
	Name    string `json:"name,string"`
	Group   string `json:"group,string"`
	Size    int    `json:"size,string"`
}

type AddEdgeMessage struct {
	Command string `json:"command,string"`
	Source  int    `json:"source,int"`
	Target  int    `json:"target,int"`
	Id      int    `json:"id,int"`
	Weight  int    `json:"weight,int"`
}

type RemoveNodeMessage struct {
	Command string `json:"command,string"`
	Id      int    `json:"id"`
}

type RemoveEdgeMessage struct {
	Command string `json:"command,string"`
	Source  int    `json:"source,int"`
	Target  int    `json:"target,int"`
	Id      int    `json:"id,int"`
}

type SetNodeNameMessage struct {
	Command string `json:"command,string"`
	Id      int    `json:"id,int"`
	Name    string `json:"name,int"`
}

func (g *Graph) BroadcastAddNode(n Node) {
	msg := AddNodeMessage{
		Command: "AddNode",
		Id:      n.Id,
		Name:    n.Name,
		Group:   n.Group,
		Size:    n.Size,
	}

	encoded, err := json.Marshal(msg)

	if err != nil {
		log.Panic("Marshaling went oh so bad: ", err)
	}

	g.wc.SendBytes(encoded)

}

func (g *Graph) BroadcastAddEdge(e Edge) {

	msg := AddEdgeMessage{
		Command: "AddEdge",
		Source:  e.Source,
		Target:  e.Target,
		Id:      e.Id,
		Weight:  e.Weight,
	}

	encoded, err := json.Marshal(msg)
	if err != nil {
		log.Panic("Marshaling went bad: ", err)
	}

	g.wc.SendBytes(encoded)
}

func (g *Graph) BroadcastRemoveNode(n Node) {
	msg := RemoveNodeMessage{
		Command: "RemoveNode",
		Id:      n.Id,
	}

	encoded, err := json.Marshal(msg)
	if err != nil {
		log.Panic("Marshaling of BroadcastRemoveNode failed, err: ", err)
	}

	g.wc.SendBytes(encoded)
}

func (g *Graph) BroadcastRemoveEdge(e Edge) {
	msg := RemoveEdgeMessage{
		Command: "RemoveEdge",
		Source:  e.Source,
		Target:  e.Target,
		Id:      e.Id,
	}

	encoded, err := json.Marshal(msg)
	if err != nil {
		log.Panic("Marshaling of BoradcastRemoveEdge failed, err: ", err)
	}

	g.wc.SendBytes(encoded)
}

func (g *Graph) BroadcastRenameNode(n Node) {
	msg := SetNodeNameMessage{
		Command: "SetNodeName",
		Id:      n.Id,
		Name:    n.Name,
	}

	encoded, err := json.Marshal(msg)

	if err != nil {
		log.Panic("Marshaling went oh so bad: ", err)
	}

	g.wc.SendBytes(encoded)
}

// Communication.go End

// Client Begin

type Client struct {
	Conn     *websocket.Conn
	Listener net.Listener
}

func NewClient(ip, port string) (c *Client, err error) {
	c = new(Client)
	address := "ws://" + ip + port
	origin := "http://" + ip

	var retryAttempts int = 0
	for {
		c.Conn, err = websocket.Dial(address, "ws", origin)
		if err != nil {
			if retryAttempts == 10 {
				return nil, err
			}
			retryAttempts++
			time.Sleep(125 * time.Millisecond)
			continue
		}
		break
	}
	return c, nil
}

func (c *Client) Send(message string) {
	msg := []byte(message)
	c.SendBytes(msg)
}

func (c *Client) SendBytes(message []byte) {
	if _, err := c.Conn.Write(message); err != nil {
		log.Panic("Message", message, " could not be sent", err)
	}

}
func (c *Client) Receive() string {

	var incoming = make([]byte, 512)
	var n int
	var err error

	if n, err = c.Conn.Read(incoming); err != nil {
		log.Fatal(err)
	}
	return string(incoming[:n])

}

// Client End

// gographer.go Begin
type Graph struct {
	Nodes map[string]*Node `json:"nodes,omitempty"`
	Edges map[string]*Edge `json:"edges,omitempty"`

	wss *WSServer
	wc  *Client
}

type Node struct {
	stringIdentifier string
	Id               int          `json:"id,int"`
	Name             string       `json:"name,string"`
	Group            string       `json:"group,string"`
	Size             int          `json:"size,int"`
	Graphics         NodeGraphics `json:"graphics,omitempty"`
}

type Edge struct {
	stringIdentifier string
	Source           int          `json:"source,int"`
	Target           int          `json:"target,int"`
	Id               int          `json:"id,int"`
	Weight           int          `json:"weight,int"`
	Graphics         EdgeGraphics `json:"graphics,omitempty"`
}

/* This is invoked first for all new connections.
 * Output current graph state before giving out updates
 */
func (g *Graph) Handler(conn *websocket.Conn) {
	b, err := json.Marshal(g)
	if err != nil {
		log.Panic("Marshaling went bad: ", err)
	}

	msg := InitGraphMessage{
		Command: "InitGraph",
		Graph:   string(b),
	}

	encoded, err := json.Marshal(msg)
	if err != nil {
		log.Panic("Marshaling went oh so bad: ", err)
	}

	conn.Write(encoded)
}

func NewG() *Graph {

	port := ":3999"
	return NewGraphAt(port)
}

// The functionality of starting wsserver on specified port
func NewGraphAt(port string) *Graph {
	graph := new(Graph)

	nodes := make(map[string]*Node)
	edges := make(map[string]*Edge)

	wsserver := New("", port)
	wsserver.SetConnectionHandler(graph)
	wsserver.Start()

	wsclient, err := NewClient("localhost", port)
	if err != nil {
		log.Fatalf("ops error creating new client")
	}

	graph.Nodes = nodes
	graph.Edges = edges
	graph.wss = wsserver
	graph.wc = wsclient

	return graph
}

func (g *Graph) ServerInfo() string {
	return g.wss.GetServerInfo()
}

// Node is uniquely identified by id
func (g *Graph) AddNode(id int, name string, group int, size int) {
	var graphics NodeGraphics
	n := &Node{Id: id, Name: name, Group: "group " + strconv.Itoa(group),
		Size: size, Graphics: graphics}

	n.stringIdentifier = fmt.Sprintf("%d", id)

	// Prevents nodes being added multiple times
	if _, alreadyAdded := g.Nodes[n.stringIdentifier]; !alreadyAdded {
		g.Nodes[n.stringIdentifier] = n
		g.BroadcastAddNode(*n)
	}
}

func (g *Graph) RemoveNode(nodeId int) {
	stringIdentifier := fmt.Sprintf("%d", nodeId)
	if node, exists := g.Nodes[stringIdentifier]; exists {
		// TODO: Remove all links associated with node.
		g.BroadcastRemoveNode(*node)
		delete(g.Nodes, stringIdentifier)
	}
}

// Add edge between Source and Target
// Edge is uniquely identified by tuple (source, target, id)
func (g *Graph) AddEdge(from, to, id, weight int) {
	var graphics EdgeGraphics

	e := &Edge{Source: from, Target: to, Id: id, Weight: weight, Graphics: graphics}

	e.stringIdentifier = fmt.Sprintf("%d-%d:%d", from, to, id)
	e.Weight = 1

	if _, exists := g.Edges[e.stringIdentifier]; !exists {
		g.Edges[e.stringIdentifier] = e
		g.BroadcastAddEdge(*e)
	}
}

func (g *Graph) RemoveEdge(from, to, id int) {
	stringIdentifier := fmt.Sprintf("%d-%d:%d", from, to, id)
	if edge, exists := g.Edges[stringIdentifier]; exists {
		g.BroadcastRemoveEdge(*edge)
		delete(g.Edges, stringIdentifier)
	}
}

func (g *Graph) RenameNode(nodeId int, name string) {
	stringIdentifier := fmt.Sprintf("%d", nodeId)
	if node, exists := g.Nodes[stringIdentifier]; exists {
		node.Name = name
		g.BroadcastRenameNode(*node)
	}
}

func (g *Graph) GetNumberOfNodes() (numberOfNodes int) {
	return len(g.Nodes)
}

type WriteToJSON struct {
	Nodes []*Node `json:"nodes,omitempty"`
	Edges []*Edge `json:"links,omitempty"`
}

// Write graph to json file
func (g *Graph) DumpJSON(filename string) {

	// Marshal
	b, err := json.Marshal(g)
	if err != nil {
		log.Panic("Marshaling of graph gone wrong")
	}

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		log.Panic("Could not create json file for graph")
	}

	// Write to file
	_, err = file.Write(b)
	if err != nil {
		log.Panic("Could not write json to file")
	}

}

// gographer.go End\"\""}}}}

// gowebsocket.go Begin

// Implementation inspired by http://gary.beagledreams.com/page/go-websocket-chat.html

type hub struct {

	// registered cnnections
	connections map[*connection]bool

	// inbound messages from connections
	broadcast chan string

	// register requests from connections
	register chan *connection

	// unregister requests from connections
	unregister chan *connection
}

var h = hub{
	connections: make(map[*connection]bool),
	broadcast:   make(chan string),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
}

func (h *hub) Run() {
	for {
		select {

		// Register new connection
		case c := <-h.register:
			h.connections[c] = true

		// Unregister connection
		case c := <-h.unregister:
			delete(h.connections, c)
			//close(c.send)

		// Broadcast message to all connections. If send buffer
		// is full unregister and close websocket connection
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					delete(h.connections, c)
					close(c.send)
					go c.conn.Close()
				}
			}

		}
	}
}

type connection struct {
	// connection
	conn *websocket.Conn

	// buffered channel of outbund messages
	send chan string
}

func (c *connection) reader(h *hub) {
	for {
		var message [1000]byte
		n, err := c.conn.Read(message[:])
		if err != nil {
			break
		}
		h.broadcast <- string(message[:n])
	}
	c.conn.Close()
}

func (c *connection) writer(h *hub) {
	for message := range c.send {
		err := websocket.Message.Send(c.conn, message)
		if err != nil {
			break
		}
	}
}

func (s *WSServer) connHandler(conn *websocket.Conn) {
	/* If a handler for new connections is registered, invoke it first
	 * so it can read to or write from the connection before we setup
	 * broadcasting
	 */
	if s.ConnHandler != nil {
		s.ConnHandler.Handler(conn)
	}

	s.Conn = &connection{send: make(chan string, 256), conn: conn}
	s.Hub.register <- s.Conn
	defer func() {
		s.Hub.unregister <- s.Conn
	}()

	go s.Conn.writer(s.Hub)
	s.Conn.reader(s.Hub)
}

type WSConnHandler interface {
	Handler(conn *websocket.Conn)
}

type WSServer struct {
	Hub         *hub
	Server      *http.Server
	Conn        *connection
	ConnHandler WSConnHandler
}

func New(ip, port string) (s *WSServer) {

	s = new(WSServer)

	s.Hub = &hub{
		connections: make(map[*connection]bool),
		broadcast:   make(chan string),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
	}

	s.Server = &http.Server{
		Addr:    ip + port,
		Handler: websocket.Handler(s.connHandler),
	}

	return s
}

/* connHandler is an optional connection handler that may be registered
 * to receive new connections before they are attached to the hub.
 * Can be nil if no connection handler is desired.
 */
func (s *WSServer) SetConnectionHandler(connHandler WSConnHandler) {
	s.ConnHandler = connHandler
}

func (s *WSServer) Start() {
	go s.Hub.Run()

	go func() {

		log.Print("handlr:", s.Server.Handler)
		http.Handle("/"+s.Server.Addr, s.Server.Handler)
		err := s.Server.ListenAndServe()

		if err != nil {
			log.Panic("Websocket server could not start:", err)
		}
	}()
	log.Print("Websocket server started successfully.")
	log.Print("Server info:", s.GetServerInfo())

}

// func customHandler(h http.Handler) http.Handler {
//     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
//         h.ServeHTTP(w,r)
//     })
// }

func (s *WSServer) GetServerInfo() string {
	return s.Server.Addr
}

// gowebsocket.go End

// graphics.go Begin

/*
   Functinality for taking control of layout of graph
*/

type NodeGraphics struct {
	Name    string `json:"name,string"`
	FGColor string `json:"fgcolor,string"`
	BGColor string `json:"bgcolor,string"`
	Shape   string `json:"shape,string"`
	X       int    `json:"x,int"`
	Y       int    `json:"y,int"`
	Height  int    `json:"height,int"`
	Width   int    `json:"width,int"`
}

type EdgeGraphics struct {
	Type  string `json:"type,string"`
	Name  string `json:"name,string"`
	Value string `json:"value,string"`
}

func (g *Graph) AddGraphicNode(id int, name string, group int, size int,
	description string, fgcolor string, bgcolor string, shape string, x int,
	y int, height int, width int) {

	graphics := NodeGraphics{description, fgcolor, bgcolor, shape, x, y, height, width}

	// TODO: Maybe clean up the below and write this and AddNode as one func

	n := &Node{Id: id, Name: name, Group: "group " + strconv.Itoa(group),
		Size: size, Graphics: graphics}

	n.stringIdentifier = fmt.Sprintf("%d", id)

	// Prevents nodes being added multiple times
	if _, alreadyAdded := g.Nodes[n.stringIdentifier]; !alreadyAdded {
		g.Nodes[n.stringIdentifier] = n
		g.BroadcastAddNode(*n)
	}
}

func (g *Graph) AddGraphicEdge(from, to, id, weight int, typ, name, value string) {

	graphics := EdgeGraphics{typ, name, value}

	e := &Edge{Source: from, Target: to, Id: id, Weight: weight, Graphics: graphics}

	e.stringIdentifier = fmt.Sprintf("%d-%d:%d", from, to, id)
	e.Weight = 1

	if _, exists := g.Edges[e.stringIdentifier]; !exists {
		g.Edges[e.stringIdentifier] = e
		g.BroadcastAddEdge(*e)
	}
}

// graphics.go End