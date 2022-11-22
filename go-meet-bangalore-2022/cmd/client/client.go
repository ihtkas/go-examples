package main

import (
	"context"
	"flag"
	"fmt"
	eventspb "github.com/ihtkas/go-examples/go-meet-bangalore-2022/gen/events"
	"github.com/ihtkas/loadgen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
	"strconv"
	"time"
)

var cl1, cl2 eventspb.EventsServiceClient

var port1 int
var port2 int

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.IntVar(&port1, "port1", 8123, "port to host gRPC server for events service")
	flag.IntVar(&port2, "port2", 8124, "port to host gRPC server for smart events service")
	flag.Parse()

	dconn, err := grpc.Dial("127.0.0.1:"+strconv.Itoa(port1), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	sdconn, err := grpc.Dial("127.0.0.1:"+strconv.Itoa(port2), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	cl1 = eventspb.NewEventsServiceClient(dconn)
	cl2 = eventspb.NewEventsServiceClient(sdconn)

	histOpt1 := loadgen.WithHistogramOpt(loadgen.DefaultHistogram("events1", 0.1, 0.1, 100))
	histOpt2 := loadgen.WithHistogramOpt(loadgen.DefaultHistogram("events2", 0.1, 0.1, 100))
	histOpt3 := loadgen.WithHistogramOpt(loadgen.DefaultHistogram("events3", 0.1, 0.1, 100))
	histOpt4 := loadgen.WithHistogramOpt(loadgen.DefaultHistogram("events4", 0.1, 0.1, 100))
	errOpt1 := loadgen.WithErrGaugeOpt(loadgen.DefaultErrGauge("events1_err"))
	errOpt2 := loadgen.WithErrGaugeOpt(loadgen.DefaultErrGauge("events2_err"))
	errOpt3 := loadgen.WithErrGaugeOpt(loadgen.DefaultErrGauge("events3_err"))
	errOpt4 := loadgen.WithErrGaugeOpt(loadgen.DefaultErrGauge("events4_err"))

	opt1 := loadgen.WithExecLimitOpt(100000)
	opt2 := loadgen.WithExecLimitOpt(100000)
	opt3 := loadgen.WithExecLimitOpt(100000)
	opt4 := loadgen.WithExecLimitOpt(100000)
	fn1 := loadgen.NewFunction(
		[]*loadgen.Stmt{loadgen.NewStmt(streamEvents(cl1, 1), 2, 5*time.Second, 5*time.Second, histOpt1, errOpt1)},
		100, opt1,
	)
	fn2 := loadgen.NewFunction(
		[]*loadgen.Stmt{loadgen.NewStmt(streamEvents(cl2, 1), 2, 5*time.Second, 5*time.Second, histOpt2, errOpt2)},
		100, opt2,
	)
	fn3 := loadgen.NewFunction(
		[]*loadgen.Stmt{loadgen.NewStmt(streamEvents(cl1, 10000), 2, 5*time.Second, 5*time.Second, histOpt3, errOpt3)},
		1, opt3,
	)
	fn4 := loadgen.NewFunction(
		[]*loadgen.Stmt{loadgen.NewStmt(streamEvents(cl2, 10000), 2, 5*time.Second, 5*time.Second, histOpt4, errOpt4)},
		1, opt4,
	)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	lg := loadgen.GenerateLoad(ctx, cancel, "load_test", []*loadgen.Function{fn1, fn2, fn3, fn4}, 5, 100, 5*time.Second, 1)
	fmt.Println("error:", lg.StartSever(":2112"))
}

func streamEvents(cl eventspb.EventsServiceClient, n int) loadgen.Evaluator {
	return func(ctx context.Context, iter uint, payload map[string]interface{}) (contRepeat bool, err error) {
		streamCtx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		publishCl, err := cl.PublishEvent(streamCtx)
		if err != nil {
			return false, err
		}
		for i := 0; i < n; i++ {
			event := make([]byte, 4)
			rand.Read(event)
			err = publishCl.Send(&eventspb.PublishEventRequest{
				Event: event,
			})
			if err != nil {
				log.Println(err)
				return false, err
			}
		}

		err = publishCl.CloseSend()
		log.Println("published events...", n)
		return true, err
	}
}
