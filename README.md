**This is an example code for the event:** https://www.facebook.com/events/2062100324084044/


# Setup 
Check out go-ethereum source to $GOPATH 
```
git clone https://github.com/ethereum/go-ethereum.git  $GOPATH/src/github.com/ethereum/go-ethereum
cd $GOPATH/src/github.com/ethereum/go-ethereum
git checkout v1.8.27
```
# Trie
In this example, we use trie library of Ethereum to write data.
## Run
```go run main.go```
# State
In this exampke, we use state library of Ethereum to create a state then add some accounts to the state.
## Run 
```go run main.go```

# Protocol
In this example, we create an application that connects to a fullnode to get block information.

## Run
Please install yaml lib: `go get gopkg.in/yaml.v2
`

Then run the app:
```go run main.go```

# nodejs-leveldb
This app is used to read data from leveldb

## Run
```
npm install
node index.js
```