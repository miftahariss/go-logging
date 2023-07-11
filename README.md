# Logger

## Installation Guidelines

### How to import private repository

1. Create and open the file with name is `.netrc` for Linux and `_netrc` for Window.
   ```sh  
   vi ~/.netrc
   ```

2. Add the line at the below into the `.netrc` or `_netrc` file.
   ```sh  
    machine gitlab.klik.doctor login [gitlab username] password [personal access token] 
   ```

- [gitlab username] : your GitLab username
- [personal access token] : personal access token from your GitLab. [please follow this guide from the GitLab docs](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#creating-a-personal-access-token)

3. Change permissions for `.netrc` or `_netrc` file.
   ```sh 
     chmod 600 ~/.netrc 
   ``` 

### Install the package from private repository

To install the keycloak package, you first need to have [Go](https://go.dev/doc/install) installed.

```sh
  # get the package from master branch
  GOPRIVATE=gitlab.klik.doctor go get gitlab.klik.doctor/platform/go-pkg/logger

  # get the package from specific tag (for example to try out the dev-tag)
  GOPRIVATE=gitlab.klik.doctor go get gitlab.klik.doctor/platform/go-pkg/logger@<tag>
```

# Quick Start
This is logger instructions for klikdokter microservice.  

The logger helper will introduce Correlation-ID (traceID) in the logs that will significantly improve troubleshooting at your application. The traceID can also be linked back to your REST-API response, for example:  
```javascript
/* 
  The correlation_id returned from the response can be linked back to your logs like this:
  [2022-11-07 11:36:34] INFO 00-1be26d3cc28eb5d9f4882df3e545a51e-3dbc120fdb4bdfe3-01 my info logs here - caller=test.go:23
  [2022-11-07 11:36:41] WARN 00-1be26d3cc28eb5d9f4882df3e545a51e-3dbc120fdb4bdfe3-01 my warning logs here - caller=test.go:32   
*/
//Sample REST-API response (with same trace-id 00-1be26d3cc28eb5d9f4882df3e545a51e-3dbc120fdb4bdfe3-01 in the correlation_id field):
{
  "meta": {
    "correlation_id": "00-1be26d3cc28eb5d9f4882df3e545a51e-3dbc120fdb4bdfe3-01",
    "code": 201000,
    "message": "Success",
    "pagination": {}
  },
  "data": {},
  "errors": {}
}  

```

Beside the traceID, there is also a log level (debug, info, warn, and error) that can be used for log filtering.  


## Quick start

### Log format

The default log format is `[Timestamp] [Level] [TraceID] [Message] - [Caller] - [Stacktrace]`

Example : 
```
[2022-11-07 10:31:53] ERROR 00-1be26d3cc28eb5d9f4882df3e545a51e-3dbc120fdb4bdfe3-01 record not found - caller=sample_product_service.go:221
[2022-11-07 11:36:34] INFO  00-75c1b9a89f6dee492f8c89c6bb4d5547-3cc9f4226287f329-01 Single session mode - caller=test.go:23
[2022-11-07 11:36:41] INFO  00-75c1b9a89f6dee492f8c89c6bb4d5547-3cc9f4226287f329-01 Continuous session mode - caller=test.go:32
[2022-11-07 11:36:41] ERROR 00-75c1b9a89f6dee492f8c89c6bb4d5547-3cc9f4226287f329-01 new errors - caller=test.go:34
```

### Log config

```tree
server:
  log:
    level: info #debug, info, warn, error.
    output: console
    file-path: ./kd-microservice.log
```

Log level :
- debug // AllowDebug allows error, warn, info and debug level log events to pass.
- info  // AllowInfo allows error, warn and info level log events to pass.
- warn  //AllowWarn allows error and warn level log events to pass.
- error // AllowError allows only error level log events to pass.

### Log level

The Log package provides methods that allow writing the log with level.

- `InfoLevel` is the default logging priority
- `WarnLevel` logs are more important than Info, but don't need individual human review.
- `ErrorLevel` logs are high-priority. If an application is running smoothly, it shouldn't generate any error-level logs.
- `DebugLevel` verbose logs that are mainly used for troubleshooting which normally being turned-off by default. Otherwise there will be too much logs written.  

It is recommended to set the default log-level to `INFO`, and sparingly use log.Info to avoid spamming the logs. Use log.Debug for logs which nature is used for troubleshooting, we can lower down the log-level when it's required.

### Log with context

Since we want to share the same traceID across one http request, we'll need to pass down the same traceID value by using **http.request.Context**, which means we'll need to pass down the said context/logger from Go kit transport layer, endpoint layer, and then to service layer.  
  
The Log package also provides support without Context.
  
P/s: please note that the TraceID only works with the Context because we need to get the trace-id in the request header or create a new one if it does not exist and put it in the context  

## Usage

### Log Without Context

```go
package main

import (
	"errors"
	"fmt"
	"gitlab.klik.doctor/platform/go-pkg/logger"
)

func main() {
	log := logger.NewLogger(
		logger.NewGoKitLog(),
	)

	//Info Level
	log.Info("starting!!!!")

	//Warn Level
	log.Warn("function deprecated!!!!")

	// Error Level
	_, err := hello("")
	if err != nil {
		log.Error(err)
	}
	// Error Level with new errors
	log.Error(errors.New("new errors"))
}

func hello(name string) (string, error) {
	if name == "" {
		return "", errors.New("empty name")
	}
	message := fmt.Sprintf("Hi, %v. Welcome!", name)
	return message, nil
}

// Output:
//[2022-11-07 10:52:53] INFO  starting!!!! - caller=main.go:17
//[2022-11-07 10:52:53] WARN  function deprecated!!!! - caller=main.go:20
//[2022-11-07 10:52:53] ERROR empty name - caller=main.go:25
//[2022-11-07 10:52:53] ERROR new errors - caller=main.go:28
```

### Log WithContext
```go
package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"gitlab.klik.doctor/platform/go-pkg/logger"
	"net/http"
	"time"
)

func main() {
	lg := logger.NewLogger(
		logger.NewGoKitLog(),
	)

	router := mux.NewRouter()
	router.Use(logger.TraceIdentifierMiddleware)
	router.HandleFunc("/single-session-mode", func(w http.ResponseWriter, r *http.Request) {
		lg.WithContext(r.Context()).Info("Single session mode")

		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	router.HandleFunc("/continuous-session-mode", func(w http.ResponseWriter, r *http.Request) {
		log := lg.WithContext(r.Context())

		log.Info("Continuous session mode")

		log.Error(errors.New("new errors"))

		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	_ = srv.ListenAndServe()
}

// Output:
//URL : single-session-mode
//[2022-11-07 11:36:34] INFO 00-75c1b9a89f6dee492f8c89c6bb4d5547-3cc9f4226287f329-01 Single session mode - caller=test.go:23

//URL : continuous-session-mode
//[2022-11-07 11:36:41] INFO  00-75c1b9a89f6dee492f8c89c6bb4d5547-3cc9f4226287f329-01 Continuous session mode - caller=test.go:32
//[2022-11-07 11:36:41] ERROR 00-75c1b9a89f6dee492f8c89c6bb4d5547-3cc9f4226287f329-01 new errors - caller=test.go:34
```

### Implement with the existing projects

- Copy the folder at the path `helper/logger` on the boilerplate to the `helper` folder in your project
- Update the log config at the `server` key, you can see the new config example at the file `config.yml.example`
```yaml
server:
  log:
    level: info #debug, error, warn, info.
    output: console
    file-path: ./kd-microservice.log
```

- From the file `main.go`, replace lines 59 to 78 by the code at the below
```go
	log := logger.NewLogger(
		logger.NewGoKitLog(),
	)
```
- Replace all the old logger (gokit-log) to new log package

Example:

```go
_ = logger.Log("message", "Connection Db Success") 
```

replace the code below

```go
log.Info("Connection Db Success")
```

- Find the `ServerOption` block in all the files in the `transport` folder and inject this one `logger.TraceIdentifier()` to the `ServerBefore` method.

Example:

from `app/api/transport/sample_product_http.go` find the block at the below

```go
	options := []httptransport.ServerOption{
        httptransport.ServerErrorLogger(logger),
        httptransport.ServerErrorEncoder(encoder.EncodeError),
        // Listener for Extract JWT from HTTP to Context
        httptransport.ServerBefore(jwt.HTTPToContext()),
    }
```
then replace the code below

```go
	options := []httptransport.ServerOption{
        httptransport.ServerErrorLogger(logger),
        httptransport.ServerErrorEncoder(encoder.EncodeError),
        // Listener for Extract JWT from HTTP to Context
        httptransport.ServerBefore(jwt.HTTPToContext(), logger.TraceIdentifier()),
    }
```

- In the `service` we will add context as the first param for all the methods

Example:

from `app/service/sample_product_service.go` 

```go
type SampleProductService interface {
	CreateSampleProduct(input request.SaveSampleProductRequest) (*response.SampleProductResponse, message.Message)
	GetSampleProduct(uid string) (*response.SampleProductResponse, message.Message)
	GetList(input request.SampleProductListRequest) ([]response.SampleProductResponse, *base.Pagination, message.Message)
	UpdateSampleProduct(uid string, input request.SaveSampleProductRequest) (*response.SampleProductResponse, message.Message)
	DeleteSampleProduct(uid string) message.Message
}
```

To 

```go
type SampleProductService interface {
    CreateSampleProduct(ctx context.Context, input request.SaveSampleProductRequest) (*response.SampleProductResponse, message.Message)
    GetSampleProduct(ctx context.Context, uid string) (*response.SampleProductResponse, message.Message)
    GetList(ctx context.Context, input request.SampleProductListRequest) ([]response.SampleProductResponse, *base.Pagination, message.Message)
    UpdateSampleProduct(ctx context.Context, uid string, input request.SaveSampleProductRequest) (*response.SampleProductResponse, message.Message)
    DeleteSampleProduct(ctx context.Context, uid string) message.Message
}
```

P/s: please update all the functions that implement the `SampleProductService` interface

- In the `Endpoints` layer we will pass the context to the service method

from `app/api/endpoint/sample_product_endpoint.go`

```go
func makeShowSampleProduct(s service.SampleProductService) endpoint.Endpoint {
	return func(ctx context.Context, rqst interface{}) (resp interface{}, err error) {
		
		//The context passed to the GetSampleProduct function as the first param
		result, msg := s.GetSampleProduct(ctx, fmt.Sprint(rqst)) 
		if msg.Code == 4000 {
			return base.SetHttpResponse(ctx, msg.Code, msg.Message, nil, nil), nil
		}

		return base.SetHttpResponse(ctx, msg.Code, msg.Message, result, nil), nil
	}
}
```

- Now we can use the log with context in the `service` layer

from `app/api/endpoint/sample_product_endpoint.go`

```go
func (s *sampleProductServiceImpl) GetSampleProduct(ctx context.Context, uid string) (*response.SampleProductResponse, message.Message) {
    //Log with context 
	log := s.logger.WithContext(ctx)
	
	result, err := s.repository.FindByUid(&uid)
	if err != nil {
		log.Error(err) // call the error log
		return nil, message.FailedMsg
	}

	if result == nil {
        log.Info(message.FailedMsg.Message) // call the info log
		return nil, message.FailedMsg
	}

	return response.SampleProductMapToResponse(*result), message.SuccessMsg
}
```

### Add the `correlation_id` to HTTP-Response

from `app/model/base/response_http.go`

find the `metaResponse` struct then add this line `CorrelationId string `json:"correlation_id"`` into that struct.

```go
type metaResponse struct {
	// CorrelationId is the response correlation_id
	//in: string
	CorrelationId string `json:"correlation_id"`
	// Code is the response code
	// example: 1000
	Code int `json:"code"`
	// Message is the response message
	// example: Success
	Message string `json:"message"`

	// Pagination of to paginate response
	// in: struct{}
	Pagination *Pagination `json:"pagination,omitempty"`
}
```

Continue to find the function `SetHttpResponse` then add the context as the first param and add the key `CorrelationId` into the metaResponse block. 

```go
func SetHttpResponse(ctx context.Context, code int, message string, result interface{}, paging *Pagination) interface{} {
	.......

	return responseHttp{
        Meta: metaResponse{
            CorrelationId: fmt.Sprint(ctx.Value(logger.TraceIDContextKey)),
            .........
        },
        .....
    }
}
```

after that find all the places in the entire project where call the `base.SetHttpResponse` function and add the context as first param.

Example:

from: `app/api/endpoint/sample_product_endpoint.go`

```go
func makeDeleteSampleProduct(s service.SampleProductService) endpoint.Endpoint {
	return func(ctx context.Context, rqst interface{}) (resp interface{}, err error) {
		msg := s.DeleteSampleProduct(ctx, fmt.Sprint(rqst))
		if msg.Code == 4000 {
			return base.SetHttpResponse(ctx, msg.Code, msg.Message, nil, nil), nil
		}

		return base.SetHttpResponse(ctx, msg.Code, msg.Message, nil, nil), nil
	}
}
```
