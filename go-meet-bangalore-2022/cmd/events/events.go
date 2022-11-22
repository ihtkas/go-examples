package main

import (
	"flag"
	eventsservice "github.com/ihtkas/go-examples/go-meet-bangalore-2022/events"
	eventspb "github.com/ihtkas/go-examples/go-meet-bangalore-2022/gen/events"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
)

var port1 int
var port2 int

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.IntVar(&port1, "port1", 8123, "port to host gRPC server for events service")
	flag.IntVar(&port2, "port2", 8124, "port to host gRPC server for smart events service")
	flag.Parse()
	eg := &errgroup.Group{}
	eg.Go(func() error {
		s := grpc.NewServer(grpc.MaxRecvMsgSize(100000000))
		eventspb.RegisterEventsServiceServer(s, &eventsservice.EventsService{})
		l, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port1))
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println("starting events server...")
		err = s.Serve(l)
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	})
	eg.Go(func() error {
		s := grpc.NewServer(grpc.MaxRecvMsgSize(100000000))
		eventspb.RegisterEventsServiceServer(s, &eventsservice.EventsService{})
		l, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port2))
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println("starting smart events server...")
		err = s.Serve(l)
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	})
	eg.Go(func() error {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":2113", nil)
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	})
	err := eg.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
