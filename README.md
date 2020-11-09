# The problem?

Developing microservices that run on cloud run has been a little bit of a challenge, usually for inner project authenticated service communication you had to worry about a couple things from within you application.

    - Service discovery
    - Auto Authentication

You usually had to have had some sort of service discovery strategy within your application code. For authentication to leverage gcp IAM to keep your service to service communication secure, you also had to make sure to ping the metadata server to get an identity token with the correct audience set.

Even though those steps where trivial, you still had to write the code or use/create some sort of library to help mitigate the code duplication between your services.

As for some rpc framework to use to auto generate your client/server/message and stubs/definitions, the main choice was gRPC.

While gRPC is quite the feature complete rpc framework that really shines in scaling out complex systems. If you are just working with a handful of services, oto might make a solid alternative for you to get you stack up and running.

# How can we improve?

To solve the service discovery and auto service authentication issue we will use [runsd](https://github.com/ahmetb/runsd)

As for more info around oto, well its better to just quote one of the creators of it, Mat Ryer [reference blog post](https://pace.dev/blog/2020/07/27/how-code-generation-wrote-our-api-and-cli.html)

> rather than obfuscate the message data (humans struggle reading binary) we wanted a more user friendly JSON API which was more familiar to developers. There are a range of great packages that allow you to expose JSON services alongside the binary gRPC ones, but they work by proxying to the binary port, rather than providing a standalone solution within themselves.
> So in our case, developer comfortability and familiarity (ours and our future API consumers) is more important than most of the technical arguments that you might make in favour of gRPC. Oto uses Go interfaces to describe the API, and in all honesty, we generally try to use Go for as much as we can.

# runsd setup

So first we will take a look at our dockerfile to square away our runsd implementation.

```dockerfile

FROM golang:1.15-alpine as builder

WORKDIR /workspace

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

ARG build_target=client

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /app ./$build_target/


# add runsd to the image
ADD https://github.com/ahmetb/runsd/releases/latest/download/runsd /runsd

# gives runsd the correct permissions
RUN chmod +x /runsd


# bare bones distroless image
FROM gcr.io/distroless/base
COPY --from=builder /app /app

# copy runsd to our final image
COPY --from=builder /runsd /runsd

# the most important part is adding runsd before we start our app
ENTRYPOINT ["/runsd", "--", "/app"]

```

# oto breakdown

First we will look at `./definitions/taco.go` that contains our service definitions for our client and server to communicate between each other.

```golang

package definitions

// TacoService contains all knowledge around consumption of tacos
type TacoService interface {

    // EatTaco handles keeping track of eating tacos
    EatTaco(request EatTacoRequest) EatTacoResponse
}

// EatTacoRequest is the request for TacoService.EatTaco.
type EatTacoRequest struct {
    // Name is your name
    Name string
    // All of the Tacos you have consumed 🌮
    Tacos []string
}

// EatTacoResponse is the response for TacoService.EatTaco.
type EatTacoResponse struct {
    // Your current taco consumption status
    TacoConsumptionStatus string
}


```

Now we can run our `./generate.sh` script to download our templates and generate our code. The templates get synced down to the `./templates/` directory and the auto generated code goes into `./generated/client` and `./generated/server/`.

Lets implement the server inside of `./server/main.go`

```golang

package main

import (
    "context"
    "errors"
    "fmt"
    "github.com/amammay/otorun/generated/server"
    "github.com/pacedotdev/oto/otohttp"
    "log"
    "net/http"
    "os"
    "time"
)

func main() {

    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }

}

func run() error {

    // creates new oto base http server
    oto := otohttp.NewServer()

    // register our taco service to the oto server
    server.RegisterTacoService(oto, &tacoService{})

    mux := http.NewServeMux()

    // map the path prefix /oto/ to the oto server and our custom middleware
    mux.Handle("/oto/", userAgentMiddleware(oto))

    // cloud run sets a port env variable that we should respect
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // create our http server and bind our handle to our mux
    httpServer := &http.Server{
        Handler:      mux,
        Addr:         ":" + port,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    }
    log.Printf("Starting server on %s", httpServer.Addr)

    return httpServer.ListenAndServe()

}

// userAgentMiddleware logs the user agent middleware for all requests that come into oto
func userAgentMiddleware(h http.Handler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        if agent := r.UserAgent(); agent != "" {
            log.Printf("info: request from %s", agent)
        }

        // send the request into oto
        h.ServeHTTP(w, r)
    }

}

// tacoService implementation
type tacoService struct {
    // TODO add database
}

// EatTaco will tell you how many tacos you have consumed
func (t *tacoService) EatTaco(ctx context.Context, request server.EatTacoRequest) (*server.EatTacoResponse, error) {
    if request.Name == "" {
        return nil, errors.New("EatTacoRequest.Name is required")
    }
    if len(request.Tacos) <= 0 {
        return nil, errors.New("EatTacoRequest.Tacos is required")
    }

    return &server.EatTacoResponse{TacoConsumptionStatus: fmt.Sprintf("%s has consumed %d tacos", request.Name, len(request.Tacos))}, nil
}


```

and now the client inside of `./client/main.go`

```golang

package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"

    "cloud.google.com/go/compute/metadata"
    "github.com/amammay/otorun/generated/client"
)

func main() {

    // if we are running on cloud run we will leverage runsd for service discovery
    tacoHost := "http://tacoserver/oto/"
    if !metadata.OnGCE() {
        tacoHost = "http://localhost:8080/oto/"
    }

    // create a new client from our auto generated client
    c := client.New(tacoHost)
    c.Debug = func(s string) {
        fmt.Println(s)
    }
    // create our taco service
    tacoService := client.NewTacoService(c)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

        // consume some tacos
        tacoResponse, err := tacoService.EatTaco(context.Background(), client.EatTacoRequest{
            Name:  "Sammy Sosa",
            Tacos: []string{"Chicken", "Chorizo"},
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        fmt.Fprint(w, tacoResponse.TacoConsumptionStatus)
    })

    // cloud run sets a port env variable that we should respect
    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }

    log.Fatal(http.ListenAndServe(":"+port, nil))

}

```

So now we can travel over to the gcp console and start to get some items deployed.

Let's start with getting our GCP projects cloud run hash. You can do that by doing the following

1. Cloud run
1. Create Service
1. Service name = hello-world -> next
1. Deploy one revision from an existing container image = gcr.io/cloudrun/hello -> next
1. Up to you if you want it to be public or private and then hit create.
1. Grab the hash from the url of the service, it will be in the format of this `https://hello-world-HASHVALUE-uc.a.run.app`

Now back to our local terminal we can get everything deployed out by running `cloudhash=HASHVALUE ./buildanddeploy.sh`

The last item we need to do is to grant the service account that is powering the client the `Cloud Run Invoker` role.

Now lets test the connection

```

curl -H \
"Authorization: Bearer $(gcloud auth print-identity-token)" \
$(gcloud run services describe tacoclient --format 'value(status.url)')

```

We should get back that `Sammy Sosa has consumed 2 tacos`.

If you poke around the logs on the tacoserver you can see our user agent middleware printing out the runsd agent that made the proxying requests.

`2020/11/08 20:50:48 info: request from runsd version=0.0.0-rc.9; Go-http-client/1.1`

Happy coding!
