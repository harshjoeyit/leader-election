package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	DefaultNumNodes = 10
	MaxUID          = 1000
)

type MessageType string

const (
	// Participate msg is sent when a node wants to participate in the leader election
	Participate MessageType = "Participate"
	// Leader msg is sent when a node has been elected as the leader
	Leader MessageType = "Leader"
)

type Message struct {
	NodeUID int // UID of the sender node
	MsgType MessageType
}

type Node struct {
	UID               int
	ClockwiseNeighbor *Node
	Ch                chan *Message // Channel to receive messages
	Ring              *Ring         // Ring to which the node belongs (needed to sync messages across nodes using waitgroup)
}

func (n *Node) Send(msg *Message) {
	n.Ch <- msg

	// For counting the total number of messages sent
	totalMessagesSent++
}

func (n *Node) Receive() {
	for {
		msg := <-n.Ch

		switch msg.MsgType {
		case Participate:
			log.Println("Node", n.UID, "received Participate message from", msg.NodeUID)

			if n.UID == msg.NodeUID {
				// If the node receives its own UID, it means that the election has completed
				// and it has been elected as the leader
				log.Println("Node", n.UID, "is the leader")

				// Set the leader of the ring
				n.Ring.Leader = n

				// Send the leader message to the next node in the ring
				n.ClockwiseNeighbor.Send(&Message{
					NodeUID: n.UID,
					MsgType: Leader,
				})

			} else if n.UID < msg.NodeUID {
				// If the UID of the sender is greater than the UID of the receiver,
				// Forward the message to the next node in the ring
				log.Println("Node", n.UID, "forwarded Participate message from", msg.NodeUID)

				n.ClockwiseNeighbor.Send(msg)

			} else {
				// Message is dropped
				log.Println("Node", n.UID, "dropped Participate message from", msg.NodeUID)
			}
		case Leader:
			log.Println("Node", n.UID, "received Leader message from", msg.NodeUID)

			if n.UID != msg.NodeUID {
				// If not the leader, forward the leader message to the next node in the ring
				log.Println("Node", n.UID, "forwarded Leader message from", msg.NodeUID)
				n.ClockwiseNeighbor.Send(msg)
			}

			// Halting the election process for the node
			log.Println("Node", n.UID, "halting the election process for itself")

			n.Ring.Wg.Done()

			return
		}
	}
}

type Ring struct {
	Nodes  []*Node
	Wg     sync.WaitGroup
	Leader *Node
}

// NewRing creates a new ring with N nodes
func NewRing(N int) *Ring {
	source := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(source)

	nodes := make([]*Node, N)
	mapUID := make(map[int]bool)

	r := &Ring{
		Nodes: nodes,
		Wg:    sync.WaitGroup{},
	}

	// Create N nodes with unique UIDs
	for i := range N {
		var UID int
		for {
			// Generate a random UID
			UID = rand.Intn(MaxUID)
			if _, ok := mapUID[UID]; !ok {
				mapUID[UID] = true
				break
			}
		}

		r.Nodes[i] = &Node{
			UID:  UID,
			Ch:   make(chan *Message, 10), // Buffered channel to avoid blocking and deadlock
			Ring: r,
		}
	}

	// Connect the nodes
	for i := range N {
		nodes[i].ClockwiseNeighbor = nodes[(i+1)%N]
	}

	return r
}

func (r *Ring) ElectLeader() {
	r.Wg.Add(len(r.Nodes))

	// When the system starts, there's no leader, every node in the ring starts
	// the election simultaneously by sending its own UID to its neighbor
	for _, node := range r.Nodes {
		// Start the receiver goroutine for each node
		go node.Receive()

		log.Println("Node", node.UID, "starting the election process")

		node.ClockwiseNeighbor.Send(&Message{
			NodeUID: node.UID,
			MsgType: Participate,
		})
	}

	// Wait for all the nodes to complete the election process
	r.Wg.Wait()
}

var totalMessagesSent int

func main() {
	N := flag.Int("nodes", DefaultNumNodes, "Number of nodes")
	flag.Parse()

	if *N < 2 {
		log.Fatal("Number of nodes should be greater than 1")
	}

	ring := NewRing(*N)

	for i := range *N {
		fmt.Printf("%d --> ", ring.Nodes[i].UID)
	}
	fmt.Printf("%d\n", ring.Nodes[0].UID)

	ring.ElectLeader()

	log.Printf("\n\nLeader: %d\n", ring.Leader.UID)
	log.Printf("Total messages sent: %d\n", totalMessagesSent)
}
