# Cover Whale Go

This is somewhat of an opinionated set of packages for Go development. By no means is this going to be forced as the only way to write Go, but it's a good starting point for most things.

## HTTP Server

Examples are [here](examples/)

### Super Simple Example

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
)

func testing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("this works!"))
}

func main() {
	s := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(9090),
	)

	routes := []cwhttp.Route{
		{
			Method: http.MethodGet,
			Path: "/testing",
			Handler: http.HandlerFunc(testing)
		},
	}

	s.RegisterSubRouter("/api/v1", routes)
	log.Fatal(s.Serve())
}
```

### Error Handlers

This library exposes an `ErrHandler` type that returns an error from the handlers. Client errors can easily be generated with the `NewClientError` function. This 
will automatically marshal the error value and return it to the caller. For example using an unauthorized error:

```go
import (
	"github.com/CoverWhale/coverwhale-go/logging"	
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
)

var logger *logging.Logger

func myHandler(w http.ResponseWriter, r *http.Request) error {

	if r.Header.Get("Authorization") == "" {
		return cwhttp.NewClientError(fmt.Errorf("unauthorized"), http.StatusUnauthorized)
	}

	.../

}

var routes = []cwhttp.Route{
	{
		Method: http.MethodGet,
		Path: "/myhandler",
		Handler: &errHandler{
			Handler: myHandler,
			Logger: logger,
		},
	},
}

```

### Handlers With a Struct Context

This library also exposes a `HandleWithContext` function. This allows for custom handlers to be created with a context value (not context.WithValue). For example 
passing a data source:

```go
import (
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
)

var db *Database

func myHandler(w http.ResponseWriter, r *http.Request, db *Database) {
	id := chi.URLParam(r, "userId")
	user := db.GetUser(id)

	json.NewEncoder(w).Encode(user)
}

var routes = []cwhttp.Route{
	{
		Method: http.MethodGet,
		Path: "/users/{userId}",
		Handler: cwhttp.HandleWithContext(myHandler, db),
	},
}

```

### Custom Handlers

To create custom handlers like above, you can just define your own type and implement the http.HandlerFunc similarly to how this library does:

```go

type myHandlerType func(http.ResponseWriter, *http.Request, MyObject)

func myCustomHandlerType(h myHandlerType, obj MyObject) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w, r, obj)
	}
}

func myHandler(w http.ResponseWriter, r *http.Request, obj MyObject) {
	obj.Do()
	.../
}

var routes = []cwhttp.Route{
	{
		Method: http.MethodGet,
		Path: "/testing",
		Handler: myHandlerType(myHandler, obj),
	},
}


```

## Logging

This might change in the future. This is a very simple, no frills logging setup that's compabile with the std library logger. 

### Example

```go
package main

import (
    "github.com/CoverWhale/coverwhale-go/logging"
)

func main() {
    logger := logging.NewLogger()

    // set level directly or with env vars
    logger.Level = logging.DebugLevel

    logger.Info("Info log")
    logger.Debug("Debug log")

    logger.Infof("logging with %s", "formatting")

    logger.Error("uh oh spaghettios")

    // Calling without the instantiated logger (relies on env vars: LOG_LEVEL=debug)
    logging.Debugf("debug level %s", "stuff")

    // You can crate a new logger with a context message 
    // Get caller returns the name of the current function
    ctx := logger.WithContext(fmt.Sprintf("function=%s", logging.GetCaller()))
    ctx.Info("another message")
}
```

### Output

Since this uses the default Go logger under the hood, you can easily set an output file instead of stdout.

```go
package main

import (
    "log"
    "github.com/CoverWhale/coverwhale-go/logging"
)

func main() {
	f, err := os.OpenFile("test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		log.Fatal("error")
	}
	defer f.Close()


	logger := logging.NewLogger()
	// use SetOutput from std logger
	logger.Logger.SetOutput(f)
}
```
