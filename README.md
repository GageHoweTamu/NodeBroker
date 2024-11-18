# NodeBroker

This is an idea I've been wanting to make for a long time. Thanks to gVisor, it's a little more within reach.

The idea is to create a platform where people can run arbitrary docker containers on a network of nodes, where both the nodes and the containers are untrusted.

gVisor is a sandbox that can run untrusted code, but the main difficulty with the node software is ensuring that gVisor is active, and encrypting all data coming in and out. It also needs to be signed somehow to ensure it hasn't been tampered with.

The master server is responsible for orchestrating communication between nodes and users, and for handling payment. It needs to be secure, and needs to be able to handle a lot of traffic.

The software/web component needs to be able to send a docker container to the server, pay, and get confirmation/output of the computation.
