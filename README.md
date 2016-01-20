apifaker can help your start a json api server in a fast way

If you would like to start a simple result json api server for testing front-end, apifaker could be a on your hand.

No need of database, just create some json file, and write two line codes, then, everthing is done, you can start implementing the happy(hope so) front-end features.

### Install
`go get github.com/Focinfi/apifaker`

### Setup

```go
  api := apifaker.NewWithApiDir("./public/api_fakers")
  // create a new ServerMux 
  mux := http.NewServeMux()
  mux.Handle("/fake_api", api)
```



