# StreamingData

gRPC server /client code that implements two streaming APIs

1. Server 
go server, in server/ folder

2. Client
nodejs client in client/ folder

3. Protobuf defenitions
schema defenitions in protocol in streamingdata/ folder


## Server
### Building and running
1. Make sure you have go, gRPC and protobuf-go installed
2. Make sure this repo is cloned to the correct path in gopath.
`$GOPATH/src/github.com/dbalan/streamingdata`
3. Generate protobuf defs by 
`cd <repo_root> && protoc -I streamingdata streamingdata/streamingdata.proto --go_out=plugins=grpc:streamingdata`

4. build and run gocode
`cd server && go build && ./server`

5. One can specify the port to listen to by passing `-p` flag.

## Client
### Running
1. Make sure nodejs and npm are installed.
2. install dependencies `cd <repo_root>/client && npm install`
3. run `node client.js --count <n> --mode 'stateless|stateful'`
