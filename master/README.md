# master

This is the master server/controller. It is run in a trusted environment, and orchestrates communication and payment between worker nodes and users.

To enable etherium transactios, make sure to have geth running:

```
sudo docker pull ethereum/client-go:latest && \
sudo docker run -it \
  -p 30303:30303 \
  -p 8545:8545 \
  ethereum/client-go \
  --http \
  --http.addr "0.0.0.0" \
  --http.port "8545" \
  --http.api "eth,net,web3"
```

run `sudo docker stop ethereum/client-go` or  ctrl-c to stop the geth container.
