## apifaker

`apifaker` can help you start a json api server in a fast way.

If you would like to start a simple json api server for testing front-end, apifaker could be a great fit.

No need of database, just create some json files, and write two line codes, then, everything is done, you can start implementing the happy(hope so) front-end features.

[Example project](https://github.com/Focinfi/youzi-server)

### Install
`go get github.com/Focinfi/apifaker`

### Usage
----
#### Add a directory

`apifaker` need a directory for containing the api json files.

#### Add api files

Rules:

1. `"resource_name"` string(required), resource name for this api routes, you can treat it as table name in a database. `apifaker` assumes that resource name is plural.

1. "`has_many`" array(optional), every element must be a string of one of the other `"resource_name"`, if a resource's `"has_many"` is not empty:
    1. The response of `GET /collention/:id` and `GET /collention` will be insert the related resources.
    2. The `DELETE /collention/:id` will also delete the related resources.

1. "`has_one`" array(optional), its rules are same as of the `"has_many`" except every element must be singular and the response of `GET /collention/:id` and `GET /collention` will be only insert the a first-found item.

1. `"columns"` array(required), columuns for resource, only support `"id" "name"`, `"type"`, `"regexp_pattern"`, `"unique"`
    1. `"id"` must be a "number" as the first cloumn.
    1. Every colmun must have at lest a `"name"` and a `"type"`.
    3. `"type"` supports: `"boolean" "number" "string" "array" "object"`, these types will be used to check every item data.
    4. `"regexp_pattern"` add regular expression for validating your string-type column, using internal `regexp` package, you could run `go doc regexp/syntax` to learn all syntax.
    5. `"unique"`: set true(default false) to specify this column should be unique.

1. `"seed"` array(optional), initial data for this resource, note that every lineitem of seeds should have columns descriped in `"columns"` array, otherwise, it will throw an non-nil error.

Here is an example for users.json

```json
{
    "resource_name": "users",
    "has_many": [
        "books"
    ],
    "columns": [
        {
            "name": "id",
            "type": "number"
        },
        {
            "name": "name",
            "type": "string",
            "regexp_pattern": "[A-z]|[0-9]",
            "unique": true
        },
        {
            "name": "phone",
            "type": "string",
            "regexp_pattern": "^132",
            "unique": true
        },
        {
            "name": "age",
            "type": "number"
        }
    ],
    "seeds": [
        {
            "id": 1,
            "name": "Frank",
            "phone": "13213213213",
            "age": 22
        },
        {
            "id": 2,
            "name": "Antony",
            "phone": "13213213211",
            "age": 22
        },
        {
            "id": 3,
            "name": "Foci",
            "phone": "13213213212",
            "age": 22
        }
    ]
}
```

And books.json

```json
{
    "resource_name": "books",
    "columns": [
        {
            "name": "id",
            "type": "number"
        },
        {
            "name": "title",
            "type": "string",
            "regexp_pattern": "[A-z]|[0-9]",
            "unique": true
        },
        {
            "name": "user_id",
            "type": "number"
        }
        
    ],
    "seeds": [
        {
            "id": 1,
            "title": "The Little Prince",
            "user_id": 1
        },
        {
            "id": 2,
            "title": "Life of Pi",
            "user_id": 2
        },
        {
            "id": 3,
            "title": "The Alchemist",
            "user_id": 1
        }
    ]
}
```

#### Creat a apifaker

```go
// if there are any errors of the directory or json file unmarshaling, err will not be nil
fakeApi, err := apifaker.NewWithApiDir("/path/to/your/fake_apis")
```

And you can use it as a http.Handler to listen and serve on a port:

```go
http.ListenAndServe(":3000", fakeApi)
```

Now almost everything is done, if we use the above users.json and books.json for the `fakerApi`, then you have a list of restful apis for users and books:

```shell
GET    /users                   
GET    /users/:id               
POST   /users                   
PUT    /users/:id               
PATCH  /users/:id               
DELETE /users/:id

GET    /books                   
GET    /books/:id               
POST   /books                   
PUT    /books/:id               
PATCH  /books/:id               
DELETE /books/:id
```

And this apis are really be able to manage the users and books resource, just like using database, what's more, it will validate every request using the rules defined in `"columns"`, in this example, rules are:

0. every request: resource with the given id must exist.
1. name of users and books must be unique and users'name must contain A-Z or 0-9.
2. phone of users must has prefix "132".
3. every POST/PATH/PUT request of books: the user with given user_id must exist.

In a word, it acts like a standard restful api server.

#### Data persistence

`apifaker` will save automatically the changes back to the json file once 24 hours and when you handlers panic something. On the other hand, you can save data manually by calling a method directly:

```go
fakeApi.SaveTofile()
```

#### Integrate other mutex

Also, you can integrate other mutex which implemnets `http.Handler` into the fakeApi, to differetiate faker api from extenal mutex, you can give fakeApi a prefix:

```go
mux := http.NewServeMux()
mux.HandleFunc("/greet", func(rw http.ResponseWriter, req *http.Request) {
rw.WriteHeader(http.StatusOK)
rw.Write([]byte("hello world"))
})

fakeApi.MountTo("/fake_api")
faker.IntegrateHandler(mux)
http.ListenAndServe("localhost:3000", fakeApi)
```

Then, `/greet` will be available, at the same time, users and books apis will change to be: 

```shell
GET     /fake_api/users                   
GET     /fake_api/users/:id               
POST    /fake_api/users                   
PUT     /fake_api/users/:id               
PATCH   /fake_api/users/:id               
DELETE  /fake_api/users/:id

GET     /fake_api/books                   
GET     /fake_api/books/:id               
POST    /fake_api/books                   
PUT     /fake_api/books/:id               
PATCH   /fake_api/books/:id               
DELETE  /fake_api/books/:id
```
