# Distributed Key-Value store

This project implements a distributed key-value store in Go, utilizing etcd's Raft algorithm for consensus and fault tolerance. The system supports CRUD operations, leader election, and log replication across nodes to ensure high availability and consistency in a distributed environment.

## Instructions to run demo

To run the demo with the nodes, run:

```sh
docker compose -f docker.compose.demo.yml up --build
```

## Development setup

To run the demo with the nodes, run:

```sh
docker compose -f docker.compose.yml up --build
```
