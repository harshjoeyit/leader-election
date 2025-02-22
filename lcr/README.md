# Prototype for Leader Election in a Ring Topology (LCR Algorithm)

## Algorithm Overview

The LCR algorithm works as follows:

1.  **Initialization:** Each node in the ring has a unique identifier (UID).
2.  **Participation:** Every node starts by sending a "Participate" message containing its UID to its clockwise neighbor.
3.  **Message Passing:** When a node receives a "Participate" message:
    *   If the received UID is greater than the node's own UID, the node forwards the message to its neighbor.
    *   If the received UID is less than the node's own UID, the node drops the message.
    *   If the received UID is equal to the node's own UID, the node declares itself the leader and sends a "Leader" message around the ring.
4.  **Leader Announcement:** When a node receives a "Leader" message, it forwards it to its neighbor, until the message returns to the leader.
5.  **Termination:** Once a node receives the Leader message, it knows who the leader is and stops participating in the election.

## Implementation Details

*   **Nodes:** Each node is represented by a `Node` struct, containing its UID, a channel for receiving messages, and a pointer to its clockwise neighbor.
*   **Ring:** The ring topology is represented by a `Ring` struct, containing a slice of nodes and a `sync.WaitGroup` to synchronize the election process and termination.
*   **Messages:** Messages are represented by a `Message` struct, containing the sender's UID and the message type.
*   **Concurrency:** Each node runs in its own goroutine, allowing for concurrent message processing.
*   **Channels:** Channels are used for communication between nodes. Buffered channels are used to prevent blocking.

## Example

To run the simulation with 5 nodes:

```bash
go run main.go -nodes 5
```