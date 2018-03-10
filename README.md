![ripple][splash]

> Structured controllers for [Echo](https://github.com/labstack/echo)

[![CircleCI](https://circleci.com/gh/nowk/ripple.svg?style=svg)][circleci]
[![Build Status](http://img.shields.io/travis/nowk/ripple.svg?style=flat-square)][travis]
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)][godoc]


## Install

    go get github.com/nowk/ripple

*gopkg.in* version

    go get gopkg.in/nowk/ripple.v0



## Usage

Create your *Controller*

    type PostsController struct {
        Index echo.HandlerFunc `ripple:"GET /"`
        Show  echo.HandlerFunc `ripple:"GET /:id"`
        ...
    }

    func (PostsController) Path() string {
        return "/posts"
    }

    func (PostsController) IndexFunc(w http.ResponseWriter, req *http.Request) {
        ...
    }

    func (PostsController) ShowFunc(c *echo.Context) error {
        ...
    }

    ...

"Group" your *Controller* onto an instance of Echo.

    echoMux := echo.New()

    ripple.Group(&PostsController{}, echoMux)

This creates a new Echo group at the `Controller#Path`, in our example `/posts`, 
with all the defined actions.

    // GET /posts     => #IndexFunc
    // GET /posts/:id => #ShowFunc

---

__Controller interface:__

Structs must implement the `Controller` interface.

    type Controller interface {
        Path() string
    }

`#Path()` is the namespace for the new Echo Group the controller will be created 
on.


__Ripple tag format:__

The ripple field tag format is a simple `<HTTP METHOD> /path`

    POST /posts

*The path must be a valid Echo path.*

The __path__ is relative to the *Controller's* defined `#Path()` method

    func (Controller) Path() string {
        return "/posts"
    }

Given the above `#Path()` any defined paths will be relative to `/posts`.

    Index  http.Handler `ripple:"GET /"`
    Update http.Handler `ripple:"PUT /:id"`
    ...

    // GET /posts
    // PUT /posts/:id


__Defining handlers:__

Handlers are defined through a *Field Name* to *\<Field Name>Func* association.

    Index  echo.HandlerFunc `ripple:"GET /"`
    Update echo.HandlerFunc `ripple:"PUT /:id"`
    ...

Will look for a method within the struct that matches `<Field Name>Func`

    func (Controller) IndexFunc(...) { ... }
    func (Controller) UpdateFunc(...) { ... }
    ...

You can also define the handler to the *Field* itself during construction.

    &PostsController{
        Index: func(w http.ResponseWriter, req *http.Request) {
            ...
        },
    }

*The associated `<Field Name>Func` will always be used over any Field assignment
if the associated method exists.*


__Defining middlewares:__

Middlewares are defined very much like handlers, with the exception of the 
ripple tag format which must be defined as such:

    Log echo.Middleware `ripple:",middleware"`

And like their handler counter parts are also looked up in the same 
`<Field Name>Func` manner.

    func (Controller) LogFunc(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWrite, req *http.Request) {
            ...
            next.ServeHTTP(w, req)
            ...
        })
    }


__Field types matter:__

The field types must be a type the associated method either *is or can be 
converted to* and must ultimately be a pattern that can be mounted onto an Echo 
Group.


## Roadmap

- [ ] Embeddable structs, especially for middleware resusability.


## License

MIT


[splash]: https://s3.amazonaws.com/assets.github.com/splash-ripple.svg
[circleci]: https://circleci.com/gh/nowk/ripple
[travis]: https://travis-ci.org/nowk/ripple
[godoc]: http://godoc.org/gopkg.in/nowk/ripple.v0
