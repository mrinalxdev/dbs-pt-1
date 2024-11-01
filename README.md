# Distributed Node Communication System

This project is a simple distributed node communication system built in Go. Each node can connect with peers, send and receive messages, and, if designated as the master node, broadcast periodic heartbeats to connected peers. 

## Features

- **Peer-to-Peer Connections**: Nodes can establish connections with other nodes using TCP.
- **Task Sending and Processing**: Nodes can send tasks to other nodes, which process them and return results.
- **Heartbeat Mechanism**: A master node can send periodic heartbeat messages to check connectivity with peers.
- **Command Line Interface (CLI)**: The node provides an interactive CLI to connect to peers, send messages, and list connected peers.

## Prerequisites

- **Go**: Make sure you have Go installed. You can download it [here](https://golang.org/dl/).

## Usage

### Starting a Node

To start a node, run:
```bash
go run main.go <node_id> <port> <is_master>
