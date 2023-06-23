# Cover Whale Go

This is somewhat of an opinionated set of packages for Go development. By no means is this going to be forced as the only way to write Go, but it's a good starting point for most things.

## Code Generation

The `cwgoctl` code generator will create an opinionated Go application for you. It will create a Makefile, Dockerfile, started cmd package, starter server package, goreleaser config, and github actions.

### Pre-Reqs
To use the targets in the Makefile you will need:
1. [Go](https://go.dev/doc/install)
2. [k3d](https://k3d.io/v5.4.6/#installation)
3. [kubectl](https://kubernetes.io/docs/tasks/tools/)
4. [Docker Desktop](https://www.docker.com/products/docker-desktop/) (or regular Docker if on Linux)

### Usage

To use the utility simply follow these steps:

1. `go mod init github.com/CoverWhale/myapp`

2. Run `cwgoctl` with your options.  For example: `cwgoctl new server --name myapp`
    > Cli documentation is [here](docs/cwgo.md)

3. Run `make tidy`
    > Pulls in all dependencies and vendors them

4. Run `make deploy-local`
    > Builds a local 3 node kubernetes cluster with registry using k3d and deploys your app to the cluster.

5. You can hit your running example app at `myapp.127.0.0.1.nip.io:8080/api/v1/testing -H "Authorization: test"`
	> A default edgedb instance is also stood up for you. The URL for the edgedb ui can be found in the output of `make deploy-local` that was ran in step 4.

6. Now you can modify the code. For example adding handlers or creating another subrouter in the `server` package.

### Extra flags

The utility has various flags to enable features that may be useful for the new app. For example: `cwgoctl new server --name myapp --enable-nats --enable-graphql`

1. `--enable-nats`
	> Sets up a NATS integration.

2. `--enable-graphql`
	> Sets up a GraphQL integration. A playground can be reached at `myapp.127.0.0.1.nip.io:8080/playground`

### EdgeDB instructions

By default, your new CoverWhale app comes with edgedb enabled. Files related to edgedb can be found under the `dbschema` folder of your new app. To access your edgedb instance, follow these steps:

1. If you haven't already, run `make tidy` and `make deploy-local` to spin up a local kubernetes deployment
	> Steps 3 & 4 from above. You will need to have docker running in order for the deploy to work

2. Port forward the edgedb service to your local machine `kubectl port-forward svc/edgedb 5656:5656`

3. Create a migration `edgedb --dsn=edgedb://localhost:5656/edgedb --tls-security=insecure migration create`

4. Run the migration `edgedb --dsn=edgedb://localhost:5656/edgedb --tls-security=insecure migrate`

5. Access the UI using the URL that was printed out when `make deploy-local` was running
	> You will need to hit this URL at least once to ensure you have properly authenticated in. Afterwards you can access it via `http://edgedb.127.0.0.1.nip.io:8080/ui`


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
		Handler: myCustomHandlerType(myHandler, obj),
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
