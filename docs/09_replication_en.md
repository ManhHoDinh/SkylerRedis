# SkylerRedis Documentation

## 9. Replication (Master-Slave)

SkylerRedis implements a basic master-slave replication model to allow for data redundancy and read scaling. In this model, a single master instance handles all write operations, and one or more slave instances receive a stream of these write commands to replicate the master's dataset.

### Roles

#### Master
-   **Default Mode**: A SkylerRedis instance starts in master mode by default.
-   **Responsibilities**:
    -   Accepts both read and write commands from clients.
    -   Keeps track of connected slaves.
    -   Propagates all successful write commands to its slaves.

#### Slave
-   **Read-Only Mode**: A slave instance is effectively read-only from a client's perspective. It does not accept write commands.
-   **Responsibilities**:
    -   Connects to a master instance upon startup.
    -   Performs a handshake to establish a replication stream.
    -   Receives and applies write commands from the master to its own local dataset.

### Configuration

#### Starting as a Master
Simply run the SkylerRedis server without any replication-specific flags:
```sh
docker run -d -p 6379:6379 --name skyler-master skyler-redis
```

#### Starting as a Slave
Use the `--replicaof` flag to specify the master's host and port. You must also run the slave on a different port.
```sh
# The <master_ip> can be 'host.docker.internal' on Docker Desktop
docker run -d -p 6380:6380 --name skyler-slave1 skyler-redis --port 6380 --replicaof <master_ip> 6379
```

### The Handshake Process

When a slave connects to a master, it initiates a handshake protocol to synchronize:

1.  **`PING`**: The slave sends a `PING` to the master to verify that the connection is alive and the master is responsive.
2.  **`REPLCONF listening-port <port>`**: The slave informs the master which port it is listening on. The master registers the slave in its internal list of replicas.
3.  **`REPLCONF capa psync2`**: The slave informs the master that it has capabilities for "PSYNC2", a more modern version of the replication protocol.
4.  **`PSYNC ? -1`**: The slave requests synchronization. The `? -1` indicates that this is a new slave requesting a full synchronization.
5.  **Master's Response (`FULLRESYNC`)**: The master responds with `+FULLRESYNC <replication_id> <offset>`. It then sends a full snapshot of its data in the form of an RDB file.
6.  **RDB Transfer**: The slave receives the RDB file and loads it into memory, overwriting any existing data. This brings its state in line with the master's at a specific point in time.
7.  **Command Propagation**: After the RDB transfer, the master begins streaming all subsequent write commands to the slave, which applies them to its local dataset.

### Command Propagation

-   After the initial handshake, any write command processed by the master (e.g., `SET`, `SADD`, `DEL`, etc.) is also sent to all connected slaves.
-   The slaves receive these commands and execute them on their own local shard instances, ensuring they stay in sync with the master.
-   Slaves periodically acknowledge their replication offset to the master via `REPLCONF ACK <offset>` commands, allowing the master to monitor replication progress.
-   **(Current State)**: The current implementation forwards write commands directly. It does not yet buffer commands for disconnected slaves or handle partial resynchronization (PSYNC2 with an offset).

This master-slave architecture provides a solid foundation for building more complex, highly available systems with SkylerRedis.
