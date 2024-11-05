Distributed File System (DFS) with Go
=====================================

This project implements a distributed file system (DFS) using Go, designed to store, retrieve, and manage files across a peer-to-peer (P2P) network. The DFS offers encrypted file storage, resilient data retrieval, and dynamic peer connections.

Features
--------

*   **Peer-to-Peer Networking**: Connects multiple peers to share and sync files within the DFS.
    
*   **File Storage & Retrieval**: Efficiently stores files using SHA-1 hashing for easy lookup and retrieval.
    
*   **Encryption**: Secures files with AES encryption, ensuring data integrity and confidentiality.
    
*   **Dynamic Peer Management**: Automatically adds, removes, and manages peers to maintain an up-to-date, resilient network.
    
*   **Fault Tolerance**: Includes robust error handling and graceful shutdown for enhanced stability.
    

Project Structure
-----------------

*   **File Server**: Manages storage and peer communications.
    
*   **Store**: Handles file reading, writing, and encryption.
    
*   **P2P Module**: Manages peer connectivity and message broadcasting.
    

How It Works
------------

1.  **Start File Server**: Each node (peer) initializes a file server to listen for incoming connections.
    
2.  **Bootstrap Network**: Nodes can connect to bootstrap nodes to discover peers and join the network.
    
3.  **Store & Retrieve Data**: Nodes can store files locally, then replicate or retrieve them from peers on demand.
    
4.  **Data Encryption**: Files are encrypted before storage and decrypted upon retrieval.
    
5.  **Graceful Handling of Network Partitions**: Manages disruptions, ensuring network resiliency and data consistency.
    

Getting Started
---------------

### Prerequisites

*   [Go](https://golang.org/doc/install)
    
*   Clone this repository:
  ```
git clone https://github.com/Pavanmanikanta98/Distributed-File-Storage.git
cd Distributed-File-Storage

```
    

### Installation

Install necessary dependencies:

```
go mod tidy
```

### Usage

To start a file server:
```
make run
```

The file server listens on a specified port and connects with bootstrap nodes for peer discovery.

### Example Commands

* **Store Data**: store a file in DFS.
  ``` s.Store("myfile.txt", reader) ```
    
* **Retrieve Data**:Retrieve a file from DFS.
  ``` r, err := s.GET("myfile.txt")```
    

### Testing

Run tests with:
```
make test 
```

Contributing
------------

Contributions are welcome! Feel free to open an issue or submit a pull request for any improvements.

License
-------

This project is licensed under the MIT License.
