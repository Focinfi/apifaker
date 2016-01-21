## apifaker

`apifaker` can help you start a json api server in a fast way

If you would like to start a simple result json api server for testing front-end, apifaker could be a on your hand.

No need of database, just create some json file, and write two line codes, then, everthing is done, you can start implementing the happy(hope so) front-end features.

### Install
`go get github.com/Focinfi/apifaker`

### Usage
----
#### fake_apis directory

`apifaker` need a directory to contains the api json files

#### Add api files

Herer is an example:

```json
{
    "Resource": "users",
    "Routes": [
        {
            "Method": "GET",
            "Path": "/user/:id",
            "Params": [
                {
                    "name": "id",
                    "desc": "User's id"
                }
            ],
            "Response": {
                "name": "Frank"
            }
        },
        {
            "Method": "GET",
            "Path": "/users",
            "Params": [],
            "Response": [
                {
                    "name": "Frank"
                },
                {
                    "name": "Frank"
                }
            ]
        }
    ]
}
```

#### Creat a apifaker

```go
  // if there are any errors of directory or json file format, err will not be nil
  fakeApi, err := apifaker.NewWithApiDir("/path/to/your/fake_apis")
```

And you can use it as a http.Handler to listen and serve on a port:

```go
  http.ListenAndServe("localhost:3000", fakeApi)
```

Now almost everthing is done, you can visit localhost:3000/users and localhost:3000/user/1 to get the json response.

#### Mount to other mutex

Also, you can compose other mutex which implemneted `http.Handler` to the fakeApi

```go
  mux := http.NewServeMux()
  mux.HandleFunc("/greet", func(rw http.ResponseWriter, req *http.Request) {
    rw.WriteHeader(http.StatusOK)
    rw.Write([]byte("hello world"))
  })

  fakeApi.MountTo("/fake_api", mux)
  http.ListenAndServe("localhost:3000", fakeApi)
```

Then, `/greet` will be available, at the same time, `/users` and `/user/1` changed to be `/fake_api/users` and `/fake_api/user/1`







