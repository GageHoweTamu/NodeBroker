# node

This is a node that forms a connection with the master server. It is responsible for completing jobs that are assigned to it by the master server. The node hardware is untrusted and containers are untrusted, so this software has the responsibility of being secure on both sides.

After a job is completed, the node sends back results and recieves payment.
