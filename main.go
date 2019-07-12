package main

import (
	"runtime"
	"flag"
	"fmt"
	"time"
	"log"
	"bytes"
	"github.com/valyala/fasthttp"
	"github.com/AubSs/fasthttplogger"
	"github.com/json-iterator/go"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Location : Meraki device observation location
type Location struct {
  Lat float64 `json:"lat,omitempty"`
  Lng float64 `json:"lng,omitempty"`
  Unc float64 `json:"unc,omitempty"`
  X []float64 `json:"x,omitempty"`
  Y []float64 `json:"y,omitempty"`
}
// ScanData : Scanning API top level data
type ScanData struct {
	Type    string     `json:"type"`
	Secret  string     `json:"secret"`
	Version string     `json:"version"`
	Data    ClientData `json:"data"`
}
// ClientData : Client data
type ClientData struct {
	ApMac        string        `json:"apMac"`
	ApFloors     []string      `json:"apFloors"`
	ApTags       []string      `json:"apTags"`
	Observations []Observation `json:"observations"`
	Tenant       string        `json:"tenant"`
}
// Observation ; Observation data
type Observation struct {
	Ssid         string       `json:"ssid"`
	Ipv4         string       `json:"ipv4"`
	Ipv6         string       `json:"ipv6"`
	SeenEpoch    float64      `json:"seenEpoch"`
	SeenTime     string       `json:"seenTime"`
	Rssi         int          `json:"rssi"`
	Manufacturer string       `json:"manufacturer"`
	Os           string       `json:"os"`
	Location     LocationData `json:"location"`
	ClientMac    string       `json:"clientMac"`
}
// LocationData : Location Data
type LocationData struct {
	Lat float64   `json:"lat"`
	X   []float64 `json:"x"`
	Lng float64   `json:"lng"`
	Unc float64   `json:"unc"`
	Y   []float64 `json:"y"`
}

type job struct {
  scanData []byte
}

var (
	loc *time.Location
	maxQueueSize = flag.Int("max_queue_size", 100, "The size of the job queue")
	maxWorkers = flag.Int("max_workers", 5, "The number of workers to start")
	port = flag.String("port", "8080", "The server port")
	bucket = flag.String("bucket", "cri.conatel.cloud", "The S3 Bucket where the data will be stored")
	location = flag.String("location", "UTC", "The time location")
	tls = flag.Bool("tls", false, "Should the server listen and serve tls")
	serverCrt = flag.String("server-tls", "server.crt", "Server TLS certificate")
	serverKey = flag.String("server-key", "server.key", "Server TLS key")
	validator = flag.String("validator", "da6a17c407bb11dfeec7392a5042be0a4cc034b6", "Meraki Sacnning API Validator")
	secret = flag.String("secret", "cjkww5rmn0001SE__2j7wztuy", "Meraki Sacnning API Secret")
)

func main() {
	// Parse flags
	flag.Parse()
	// Configure concurrent jobs
  runtime.GOMAXPROCS(runtime.NumCPU())
  fmt.Println("GOMAXPROCS =", runtime.NumCPU())
  fmt.Println("maxWorkers =", *maxWorkers)
  fmt.Println("maxQueueSize =", *maxQueueSize)
  fmt.Println("port =", *port)
	// Set the time location
	var err error
  loc, err = time.LoadLocation(*location)
  if err != nil {
    panic(err)
  }
	// Create job channel
  jobs := make(chan job, *maxQueueSize)
	// Create an AWS Session and an s3 service
  sess, err := session.NewSession(&aws.Config{
    Region: aws.String("us-east-1"),
  })
  if err != nil {
		fmt.Printf("Could not configure the AWS Go SDK: %v\n", err)
  }
	svc := s3.New(sess)
	// Create the worker handler
	processBody := createProcessBody(svc)
	// Create workers
  for index := 1; index <= *maxWorkers; index++ {
    fmt.Println("Starting worker #", index)
    go func(index int) {
      for job := range jobs {
        processBody(index, job)
      }
    }(index)
  }
	// Fasthttp configuration
	handler := requestHandler(jobs)
	if err := fasthttp.ListenAndServe("0.0.0.0:" + *port, fasthttplogger.Combined(handler)); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func requestHandler(jobs chan job) func (ctx *fasthttp.RequestCtx) {
	return func (ctx *fasthttp.RequestCtx) {
		path := string(ctx.URI().Path())
		method := string(ctx.Method())
		if method == "GET" {
			if path == "/loaderio-e482326fbb627da5b2ce44f66c07fee0/" {
				ctx.SetContentType("text/plain; charset=utf8")
				fmt.Fprintf(ctx, "loaderio-e482326fbb627da5b2ce44f66c07fee0")
				return
			}
			if path == "/" {
				getValidator(ctx)
				return
			}
			if path == "/healthz" {
				getHealthz(ctx)
				return
			}
		}
		if method == "POST" {
			if path == "/" {
				getData(ctx, jobs)
				return
			}
		}
		ctx.Response.SetStatusCode(404)
		ctx.SetContentType("text/plain; charset=utf8")
		fmt.Fprintf(ctx, "404 - Not Found")
	}
}

func createProcessBody(svc *s3.S3) func(id int, j job) {
	return func (id int, j job) {
		var scanData ScanData
		err := json.Unmarshal(j.scanData, &scanData)
		if err != nil {
			fmt.Println("Can't decode body")
			return
		}
		if *secret != scanData.Secret {
			fmt.Println("Invalid secret", scanData.Secret)
			return
		}
		start := time.Now()
		data := scanData.Data
		now := time.Now().In(loc).Format(time.RFC3339)
		key := now + "-" + data.ApMac + ".json"
		data.Tenant = "tata"
		dataBytes, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		input := &s3.PutObjectInput{
			Body: bytes.NewReader(dataBytes),
			Bucket: aws.String(*bucket),
			Key: aws.String(key),
		}
		_, err = svc.PutObject(input)
		if err != nil {
			panic(err)
		}
		fmt.Println("Saved", key, "to S3 in", time.Since(start))
	}
}

func getValidator (ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/plain; charset=utf8")
	fmt.Fprintf(ctx, *validator)
} 

func getHealthz (ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(204)
}

func getData (ctx *fasthttp.RequestCtx, jobs chan job) {
	scanData := ctx.Request.Body()
	go func() {
    jobs <- job{scanData}
  }()
  // Render success
	ctx.Response.SetStatusCode(204)
}