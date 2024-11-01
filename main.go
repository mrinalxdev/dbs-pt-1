package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	From    int    `json:"from"`
}

type Node struct {
	ID      int
	IsMaster bool
	Peers   map[int]string
	conn    map[int]net.Conn
	mutex   sync.RWMutex
}

func NewNode(id int, isMaster bool) *Node {
	return &Node{
		ID:      id,
		IsMaster: isMaster,
		Peers:   make(map[int]string),
		conn:    make(map[int]net.Conn),
		mutex:   sync.RWMutex{},
	}
}

func (n *Node) Start(port int) {
	// Start listening for connections
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to start node %d: %v", n.ID, err)
	}
	defer listener.Close()

	fmt.Printf("Node %d started on port %d (Master: %v)\n", n.ID, port, n.IsMaster)

	// Accept connections in a goroutine
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept connection: %v", err)
				continue
			}

			go n.handleConnection(conn)
		}
	}()

	// If master, start heartbeat
	if n.IsMaster {
		go n.sendHeartbeats()
	}

	// Start command line interface
	n.startCLI()
}

func (n *Node) handleConnection(conn net.Conn) {
	decoder := json.NewDecoder(conn)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			return
		}

		switch msg.Type {
		case "heartbeat":
			if !n.IsMaster {
				fmt.Printf("Heartbeat received from master (Node %d)\n", msg.From)
			}
		case "task":
			fmt.Printf("Task received from Node %d: %s\n", msg.From, msg.Content)
			// Simulate task processing
			time.Sleep(time.Second)
			n.sendMessage(msg.From, Message{
				Type:    "result",
				Content: fmt.Sprintf("Processed: %s", msg.Content),
				From:    n.ID,
			})
		case "result":
			fmt.Printf("Result received from Node %d: %s\n", msg.From, msg.Content)
		}
	}
}

func (n *Node) connectToPeer(id int, address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	n.mutex.Lock()
	n.Peers[id] = address
	n.conn[id] = conn
	n.mutex.Unlock()

	return nil
}

func (n *Node) sendMessage(targetID int, msg Message) {
	n.mutex.RLock()
	conn, exists := n.conn[targetID]
	n.mutex.RUnlock()

	if !exists {
		log.Printf("No connection to node %d", targetID)
		return
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		log.Printf("Failed to send message to node %d: %v", targetID, err)
	}
}

func (n *Node) sendHeartbeats() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		n.mutex.RLock()
		for id := range n.Peers {
			n.sendMessage(id, Message{
				Type: "heartbeat",
				From: n.ID,
			})
		}
		n.mutex.RUnlock()
	}
}

func (n *Node) startCLI() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Node %d > ", n.ID)
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		parts := strings.Split(cmd, " ")

		switch parts[0] {
		case "connect":
			if len(parts) != 3 {
				fmt.Println("Usage: connect <node_id> <address>")
				continue
			}
			id, _ := strconv.Atoi(parts[1])
			err := n.connectToPeer(id, parts[2])
			if err != nil {
				fmt.Printf("Failed to connect: %v\n", err)
			} else {
				fmt.Printf("Connected to Node %d\n", id)
			}

		case "send":
			if len(parts) < 3 {
				fmt.Println("Usage: send <node_id> <message>")
				continue
			}
			targetID, _ := strconv.Atoi(parts[1])
			content := strings.Join(parts[2:], " ")
			n.sendMessage(targetID, Message{
				Type:    "task",
				Content: content,
				From:    n.ID,
			})

		case "list":
			fmt.Println("Connected peers:")
			n.mutex.RLock()
			for id, addr := range n.Peers {
				fmt.Printf("Node %d: %s\n", id, addr)
			}
			n.mutex.RUnlock()

		case "exit":
			return

		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  connect <node_id> <address> - Connect to another node")
			fmt.Println("  send <node_id> <message>    - Send a message to a node")
			fmt.Println("  list                        - List connected peers")
			fmt.Println("  help                        - Show this help")
			fmt.Println("  exit                        - Exit the program")

		default:
			fmt.Println("Unknown command. Type 'help' for available commands")
		}
	}
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run main.go <node_id> <port> <is_master>")
		os.Exit(1)
	}

	nodeID, _ := strconv.Atoi(os.Args[1])
	port, _ := strconv.Atoi(os.Args[2])
	isMaster, _ := strconv.ParseBool(os.Args[3])

	node := NewNode(nodeID, isMaster)
	node.Start(port)
}