User Guide
==========

## Set Up

Make sure you download the required dependencies: <br />

```
$ go get github.com/dgrijalva/jwt-go
$ go get github.com/gin-gonic/gin
$ go get go.mongodb.org/mongo-driver/bson/primitive
$ go get go.nanomsg.org/mangos
$ go get github.com/gorilla/websocket
```
Or simply:

```
$ go get .
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

- Login

```
$ curl -u username:password http://localhost:8080/login
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjIzNDQyNzksInVzZXIiOiJ1c2VybmFtZS5wYXNzd29yZCJ9.ARAAyNEec3xq5IrjOnBrMsKW3-ptNR9-WruAKh3XWQ4",
    "user": "username"
}
```

- System Status 

```
$ curl --location --request GET 'http://localhost:8080/status' --header 'Authorization: Bearer <ACCESS_TOKEN>'
{
    "active_workloads": 0,
    "server_time": "2021-05-29 18:03:06",
    "system_name": "Distributed Parallel Image Processing (DPIP) System"
}
```

- Create a workload

```
$ curl --location --request POST 'http://localhost:8080/workloads' --header 'Authorization: Bearer <ACCESS_TOKEN>' --form 'workload_name="myworkload"' --form 'filter="grayscale"'
{
    "filter": "grayscale",
    "filtered_images": [],
    "running_jobs": 0,
    "status": "running",
    "workload_id": "0",
    "workload_name": "myworkload"
}
```

- Check workload details by workload_id 

```
$ curl --location --request GET 'http://localhost:8080/workloads/{workload_id}' --header 'Authorization: Bearer <ACCESS_TOKEN>'
{
    "filter": "grayscale",
    "filtered_images": [],
    "running_jobs": 0,
    "status": "running",
    "workload_id": "0",
    "workload_name": "myworkload"
}
```

- Upload an image (in this case `bunny.jpg`) to a specific workload

```
$ curl --location --request POST 'http://localhost:8080/images' --header 'Authorization: Bearer <ACCESS_TOKEN>' --form 'data=@"./bunny.jpg"' --form 'workload_id="0"'
{
    "image_id": "o1_myworkload",
    "type": "original",
    "workload_id": "0"
}
```

*Note:* You can use our [test script stress_test.py](https://github.com/the-other-mariana/dc-final/blob/main/stress_test.py) to do this for the frames folder images by typing:

```
$ python3 stress_test.py -action push -workload-id 0 -token <ACCESS_TOKEN> -frames-path frames
```

- Check workload details by workload_id

```
$ curl --location --request GET 'http://localhost:8080/workloads/{workload_id}' --header 'Authorization: Bearer <ACCESS_TOKEN>'
{
    "filter": "grayscale",
    "filtered_images": [
        "f1_myworkload"
    ],
    "running_jobs": 0,
    "status": "completed",
    "workload_id": "0",
    "workload_name": "myworkload"
}
```

- Download an image by its image_id (it's their filename as well, but w/o the extension)

```
$ curl --location --request GET 'http://localhost:8080/images/{image_id}' --header 'Authorization: Bearer <ACCESS_TOKEN>'
```

*Note:* the resulting filtered images of `myworkload` (or any workload name) are in ./public/results/workload_name/ and you can join the results in a video using another [script video_utils.py](https://github.com/the-other-mariana/dc-final/blob/main/video_utils.py) by typing:

```
python3 video_utils.py -action join -workload_name myworkload filtered.mp4 public/results/myworkload
```

- Logout

```
$ curl --location --request DELETE 'http://localhost:8080/logout' --header 'Authorization: Bearer <ACCESS_TOKEN>'
{
    "logout_message": "Bye username, your token has been revoked"
}
```
