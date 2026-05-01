[Русская версия](README.ru.md)

# InSync

InSync is a file synchronization tool for a local network, built around a three-snapshot model: base, local, and remote. The project uses three-way merge to compare states, detect conflicts, and apply changes in a predictable order.

InSync is designed for synchronizing file trees over the network in a REPL format. Nodes are discovered through mDNS, data exchange is handled via gRPC, and the connection is protected by certificate-based mutual authentication.

---

## Features

- Three-way synchronization using the base / local / remote scheme
- Bidirectional sync
- Conflict detection when states diverge
- Deterministic change merging
- A contract-based model for paths and tree structure
- Strict normalized path representation
- An extensible command and transport architecture

---

## Architecture

InSync is built around the concept of three snapshots:

- **Base** — the previous synchronized state
- **Local** — the current state on the local node
- **Remote** — the current state on the remote node

### System layers

**REPL layer**  
The user interacts with the application through an interactive console and enters commands as needed.

**Transport layer**  
gRPC is used for data exchange between nodes.

**Discovery layer**  
mDNS is used to discover nodes in the local network.

**Security layer**  
The connection is protected by mutual authentication: both client and server present certificates.

**Hashing layer**  
BLAKE3 is used for computing hashes.

### Implementation details

The contract-based model makes it possible to traverse the tree in lexicographic order and skip unchanged subtrees. This reduces unnecessary work and makes the system’s behavior more predictable.

---

## Core ideas

**Contract first**  
All paths must be normalized, and the tree structure must be predefined and validated. This simplifies state handling and makes the synchronization algorithm deterministic.

**Connection security**  
Node interaction uses certificate-based mutual authentication. This reduces the risk of connecting to an unverified node.

**Determinism**  
The same input data must always produce the same result.

**Simplicity and lightweight design over universality**  
Some complex cases are intentionally not supported, including symlinks and virtual file systems.

---

## Requirements

The following are required to build and run the project:

- Go
- certificate support for both client and server nodes
- the ability to run nodes in a local network
- Windows or another environment compatible with the certificate generation scripts provided in the repository

---

## Installation

```bash
git clone https://github.com/1ight181/InSync
cd InSync
go build -o insync ./cmd/insync
```

If you are using the Windows certificate generation script, it is convenient to run it separately from the `certsgen` directory.

---

## Certificate preparation

Before starting the nodes, you need to prepare a CA certificate and then issue and sign certificates for the server and client.

If the repository uses the ready-made batch scripts from `certsgen`, first create the CA, then generate the node certificates based on it.

Procedure:

1. Generate the CA certificate
2. Generate the server certificate
3. Generate the client certificate
4. Make sure both nodes use a consistent certificate set

---

## Usage

### 1. Start the REPL

```
insync
```

After startup, the user enters interactive mode and can execute commands sequentially.

---

### 2. Discover nodes

```
insync> nodes
```

This command prints the list of available nodes in the local network.

---

### 3. Connect to a node

```
insync> connect nodeNameExample
```

After connection, state exchange and synchronization control become available.

---

### 4. Add a root

A root directory is created on each node, and its contents will be synchronized.

On the local side:

```
insync> add-root exampleRoot "/path/to/example-root"
```

On the remote side:

```
insync> add-root exampleRoot "/path/to/remote-root"
```

---

### 5. Initialize the base snapshot

After adding a root, you can initialize the base state:

```
insync> init exampleRoot
```

---

### 6. Check changes without applying them

```
insync> dry-run exampleRoot
```

This command shows the list of planned changes and possible conflicts without applying any modifications.

---

### 7. Apply synchronization

```
insync> sync exampleRoot
```

This command applies all required changes sequentially. If conflicts are found, they are shown to the user for resolution.

---

### 8. Assign an alias to a node

```
insync> set-alias nodeNameExample aliasExample
```

The alias is displayed next to the node name and makes working in the REPL easier.

---

## REPL commands

- `nodes` — show available nodes
- `connect <nodeName>` — connect to a node
- `current-node` — show the currently connected node
- `add-root <rootName> <path>` — add a root directory for synchronization
- `remove-root <rootName>` — remove a root directory from synchronization
- `init <rootName>` — create a base snapshot for selected root
- `dry-run <rootName>` — show the synchronization plan without applying it
- `sync <rootName>` — apply changes
- `set-alias <nodeName> <alias>` — assign an alias to a node
- `remove-alias <alias>` — remove an assigned alias

---

## Limitations

- Symlinks are not supported
- Virtual file systems are not supported

---

## Roadmap

- [x]  Create a custom gRPC resolver for mDNS
- [x]  Switch to a CGO-free sqlite driver
- [x]  Migrate from zeroconf to mdns
- [ ]  Add support for selecting a default conflict resolution policy
- [ ]  Increase the level of parallelism
- [ ]  Raise test coverage to 90%
- [ ]  Add daemon mode
- [ ]  Move to Bubble Tea
- [ ]  Move to rsync algorithms

---

## Repository structure

```
├── certsgen  
├── cmd  
├── config  
├── internal  
├── proto  
├── go.mod  
└── go.sum
```
## License

[MIT License](https://github.com/1ight181/InSync/blob/main/LICENSE) - Copyright (c) 2026 1ight181

## Community & support

Found a bug?  - [GitHub Issues](https://github.com/1ight181/InSync/issues)

Email: danil.odinzov181@gmail.com