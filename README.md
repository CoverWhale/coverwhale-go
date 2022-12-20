# Cover Whale Go

This is somewhat of an opinionated set of packages for Go development. By no means is this going to be forced as the only way to write Go, but it's a good starting point for most things.

## HTTP Server

Example is [here](examples/http_server/main.go)

### Super Simple Example

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/CoverWhale/coverwhale-go/logging"
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
)

var routes = []cwhttp.Route{
	{
		Method:  http.MethodGet,
		Path:    "/testing",
		Handler: testing,
	},
}

func testing(w http.ResponseWriter, r *http.Request) error {
	w.Write([]byte("this works!"))
	return nil
}


func main() {
	s := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(9090),
	)

	s.RegisterSubRouter("/api/v1", routes)
	log.Fatal(s.Serve())
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
