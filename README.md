## apifaker

`apifaker` can help you start a json api server in a fast way

If you would like to start a simple json api server for testing front-end, apifaker could be a on your hand.

No need of database, just create some json file, and write two line codes, then, everthing is done, you can start implementing the happy(hope so) front-end features.

### Install
`go get github.com/Focinfi/apifaker`

### Usage
----
#### Add a directory

`apifaker` need a directory to contain the api json files

#### Add api files

Rules:

1. `"resource_name"` string, resource name for this api route, you can see it as a table name when using database.

2. `"columns"` array, columuns for resource, Each column should at least has `"name"` and `"type"`.

3. `"seed"` array, lineitems for this resource, note that every lineitem of seeds should has columns descriped in `"columns"` array, otherwise, it will throw an non-nil error.

Here is an example for users.json

```json
{
    "resource_name": "users",
    "columns": [
        {
            "name": "name",
            "type": "string"
        },
        {
            "name": "phone",
            "type": "string"
        },
        {
            "name": "age",
            "type": "number"
        }
    ],
    "seeds": [
        {
            "name": "Frank",
            "phone": "13213213213",
            "age": 22
        },
        {
            "name": "Antony",
            "phone": "13213213211",
            "age": 22
        },
        {
            "name": "Foci",
            "phone": "13213213212",
            "age": 22
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

Now almost everthing is done, let's assume that we use the above example users.json file for the `fakerApi`, then you have a list of restful apis for users:

```shell
GET    /users                   
GET    /users/:id               
POST   /users                   
PUT    /users/:id               
PATCH  /users/:id               
DELETE /users/:id
```

And this apis are really be able to manage the users resource, just like using database.

#### Mount to other mux

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

Then, `/greet` will be available, at the same time, users apis will change to be: 

```shell
GET     /fake_api/users                   
GET     /fake_api/users/:id               
POST    /fake_api/users                   
PUT     /fake_api/users/:id               
PATCH   /fake_api/users/:id               
DELETE  /fake_api/users/:id
```

### TODO
0. Writing data into json files for persistence.
1. Add type checking for POST PUT PATCH.
2. Add relationship for deferent resources. like one2one, one2many, many2many.






