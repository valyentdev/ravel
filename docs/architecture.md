# Architecture

A Ravel cluster is composed servers and agents. The server is responsible for scheduling workloads on the cluster and the agents are responsible for running the workloads.

## Agent

The Ravel Agent is responsible for managing ravel Machines assigned to one host in the Ravel cluster. It is responsible to run the workloads with the Ravel Runtime. It answers to reservations requests broadcasted by the Ravel Server on NATS.

## Server

The Ravel Server is responsible for accepting API requests to schedule workloads on the cluster. The Ravel Server stores its state in a Postgres database and uses HTTP and NATS to communicate with the agents. It schedules workloads by broadcasting reservations requests to the agents; then it sorts the answers and assigns the workloads to the best agent.