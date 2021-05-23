User Guide
==========

## Set Up

Make sure you download the required dependencies: <br />

```
$ go get github.com/dgrijalva/jwt-go
$ go get github.com/gin-gonic/gin
$ go get go.mongodb.org/mongo-driver/bson/primitive
$ go get go.nanomsg.org/mangos
$ go get github.com/CodersSquad/dc-labs/challenges/third-partial/proto
$ go get github.com/gorilla/websocket
```

### Terminal 1: The Server

Open a terminal in the root directory, and type: <br />

```
$ export GO111MODULE=off
$ go run main.go
```

### Terminal 2: Create Workers

Open another terminal in the root directory, and type: <br />

```
$ go run ./worker/main.go --controller tcp://localhost:40899 --worker-name worker0 --tags tag1,tag2
```
*Note: Make sure that tcp://localhost:40899 is the controller address*

### Terminal 3: Requests

Open a third terminal and type: <br />

```
$ curl -u username:password http://localhost:8080/login

$ curl -H "Authorization: Bearer <ACCESS_TOKEN>" http://localhost:8080/status/<worker-name>

$ curl -H "Authorization: Bearer <ACCESS_TOKEN>" http://localhost:8080/workloads/test
{"Job ID":0,"Result":"Done by worker0","Status":"Scheduling","Workload":"test"}

$ curl -H "Authorization: Bearer <ACCESS_TOKEN>" http://localhost:8080/workloads/test
{"Job ID":1,"Result":"Done by worker0","Status":"Scheduling","Workload":"test"}
```
