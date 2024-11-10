# Consensus
Distributed Mutual Exclusion of a shared resource using the Ricart-Agrawala algorithm.

## Usage

### Make evnironment file of ports

Create a file named ``.env`` in the root directory with the wanted amount of ports.
```sh
PORTS=50051,50052,50053
```
_Each port has to be seperated with `,` after each port_
### Run the program

Run the program from root:
```sh
go run src/main.go
```
_Note:_ you have 5 sec to open the nessesarry amount of ports from the `.env` file before the peer nodes begins to access the critical section.


## Misc.
### Generate GRPC

```sh
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative consensus/consensus.proto
```

Note:
Needs to be ran from the root.