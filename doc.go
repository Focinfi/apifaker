// Package apifaker can help you start a json api server in a fast way
//
// If you would like to start a simple json api server for testing front-end, apifaker could be a on your hand.
// No need of database, just create some json file, and write two line codes, then, everthing is done, you can start implementing the happy(hope so) front-end features.
//
// Usage
//
// 1. Add a directory
//
// apifaker need a directory to contain the api json files
//
// 2. Add api files
//
// Rules:
//
// 1. `"resource_name"` string(required), resource name for this api route, you can take it as a table name when using database. As a result
//
// 2. `"columns"` array(required), columuns for resource, support `"id" "name"`, `"type"`, `"regexp_pattern"`, `"unique"`
//     1. `"id" "name"` and `"type"` are required.
//     2. `"id"` must be a "number" as the first cloumn.
//     3. `"type"` supports: `"boolean" "number" "string" "array" "object"`, these types will be used to check every item data.
//     4. `"regexp_pattern"` add regular expression for your string-type column, using internal `regexp` package, you could run `go doc regexp/syntax` to learn all syntax.
//     5. `"unique"`: set true(default false) to specify this column should be unique.
//
// 3. `"current_id"` number(required), record the the most recently item's id.
//
// Here is an example for users.json
//
// {
//     "resource_name": "users",
//     "columns": [
//         {
//             "name": "id",
//             "type": "number"
//         },
//         {
//             "name": "name",
//             "type": "string",
//             "regexp_pattern": "[A-z]|[0-9]",
//             "unique": true
//         },
//         {
//             "name": "phone",
//             "type": "string",
//             "regexp_pattern": "132.*",
//             "unique": true
//         },
//         {
//             "name": "age",
//             "type": "number"
//         }
//     ],
//     "current_id": 3,
//     "seeds": [
//         {
//             "id": 1,
//             "name": "Frank",
//             "phone": "13213213213",
//             "age": 22
//         },
//         {
//             "id": 2,
//             "name": "Antony",
//             "phone": "13213213211",
//             "age": 22
//         },
//         {
//             "id": 3,
//             "name": "Foci",
//             "phone": "13213213212",
//             "age": 22
//         }
//     ]
// }
//
//
// 3. Creat a apifaker
//
// if there are any errors of directory or json file format, err will not be nil
// fakeApi, err := apifaker.NewWithApiDir("/path/to/your/fake_apis")
//
// And you can use it as a http.Handler to listen and serve on a port:
//
//   http.ListenAndServe("localhost:3000", fakeApi)
//
// Now almost everthing is done, let's assume that we use the above example users.json file for the fakerApi, then you have a list of restful apis for users:
//
//   GET    /users
//   GET    /users/:id
//   POST   /users
//   PUT    /users/:id
//   PATCH  /users/:id
//   DELETE /users/:id
//
// And this apis are really be able to manage the users resource, just like using database.
//
// 4. Data persistence
//
// apifaker will save automatically the changes back to the json file once 24 hours and when you handlers panic something. On the other hand, you can save data manually:
//
//   fakeApi.SaveTofile()
//
// 5. Mount to other mux
//
// Also, you can compose other mutex which implemneted http.Handler to the fakeApi
//
//   mux := http.NewServeMux()
//   mux.HandleFunc("/greet", func(rw http.ResponseWriter, req *http.Request) {
//     rw.WriteHeader(http.StatusOK)
//     rw.Write([]byte("hello world"))
//   })
//
//   fakeApi.MountTo("/fake_api", mux)
//   http.ListenAndServe("localhost:3000", fakeApi)
//
// Then, /greet will be available, at the same time, users apis will change to be:
//
//   GET     /fake_api/users
//   GET     /fake_api/users/:id
//   POST    /fake_api/users
//   PUT     /fake_api/users/:id
//   PATCH   /fake_api/users/:id
//   DELETE  /fake_api/users/:id
//
package apifaker
