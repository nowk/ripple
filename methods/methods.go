package methods

// Map maps all HTTP methods to their echo func method equivalent
var Map = map[string]string{
	"GET":     "Get",
	"POST":    "Post",
	"PUT":     "Put",
	"PATCH":   "Patch",
	"DELETE":  "Delete",
	"HEAD":    "Head",
	"OPTIONS": "Options",
	"CONNECT": "Connect",
	"TRACE":   "Trace",
}
