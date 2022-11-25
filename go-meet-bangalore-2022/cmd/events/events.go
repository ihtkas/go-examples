package main

import (
	"cloud.google.com/go/profiler"
	"flag"
	eventsservice "github.com/ihtkas/go-examples/go-meet-bangalore-2022/events"
	eventspb "github.com/ihtkas/go-examples/go-meet-bangalore-2022/gen/events"
	"github.com/ihtkas/go-examples/go-meet-bangalore-2022/interceptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
)

var port int
var metricsPort int

var enableStackDriver bool
var interceptorOpt string

const (
	InterceptorV1 = "v1"
	InterceptorV2 = "v2"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.IntVar(&port, "port", 8123, "port to host gRPC server for events service")
	flag.IntVar(&metricsPort, "metricsPort", 2113, "port to host metrics server for events service")
	flag.BoolVar(&enableStackDriver, "enableStackDriver", false, "enable continuous profile monitoring using google cloud profiler")
	flag.StringVar(&interceptorOpt, "interceptor", InterceptorV1, "Version of stream interceptor to use. Choices: v1, v2. Default is v1")

	flag.Parse()
	eg := &errgroup.Group{}
	eg.Go(func() error {
		var streamInterceptor grpc.StreamServerInterceptor
		switch interceptorOpt {
		case InterceptorV1:
			streamInterceptor = interceptor.StreamInterceptor()
		case InterceptorV2:
			streamInterceptor = interceptor.StreamInterceptorV2()
		default:
			log.Fatal("Invalid interceptor. Pass v1 or v2")
		}
		s := grpc.NewServer(grpc.StreamInterceptor(streamInterceptor))
		eventspb.RegisterEventsServiceServer(s, &eventsservice.EventsService{})
		l, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println("starting events server...", port)
		err = s.Serve(l)
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	})
	eg.Go(func() error {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":"+strconv.Itoa(metricsPort), nil)
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	})
	if enableStackDriver {
		eg.Go(func() error {
			cfg := profiler.Config{
				Service:        "events" + strconv.Itoa(port),
				ServiceVersion: "1.0.0",
				ProjectID:      "crucial-guard-369408",
			}

			// Profiler initialization, best done as early as possible.
			if err := profiler.Start(cfg); err != nil {
				log.Println(err)
				return err
			}
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
