Grade Machine
=============

Built to help distribute project grades and comments to my ASU VCD students in an efficient but still reasonably obfucated way. Because have you used Blackboard?

The project was an excercise in building a simple, performant app from scratch that used an AngularJS frontend and a Go backend API.

[View the demo](http://astockwell.github.io/grade-machine) (uses a mock backend).

### Angular tid-bits

- Provides as-you-type dumb-ish validations of the form fields (using native Angular bindings)

### Go tid-bits

- Built the json backend server from only Golang standard libs
- Grades/roster held in in-memory struct and seeded via a json file (writable and multi-tenant backends were out of scope)
- Template views and grades json seed file used automatic hot-reloading
- Request "authentication" via a matching of Student ID and Last Name
