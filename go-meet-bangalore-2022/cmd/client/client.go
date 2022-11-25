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

var cl eventspb.EventsServiceClient

var port int
var metricsPort int

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.IntVar(&port, "port", 8123, "port to host gRPC server for events service. default is 8123")
	flag.IntVar(&metricsPort, "metricsPort", 2112, "port to host metrics server for events service. default is 2112")
	flag.Parse()

	dconn, err := grpc.Dial("127.0.0.1:"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	cl = eventspb.NewEventsServiceClient(dconn)

	histOpt1 := loadgen.WithHistogramOpt(loadgen.DefaultHistogram("events1", 0.1, 0.1, 100))
	histOpt3 := loadgen.WithHistogramOpt(loadgen.DefaultHistogram("events3", 0.1, 0.1, 100))
	errOpt1 := loadgen.WithErrGaugeOpt(loadgen.DefaultErrGauge("events1_err"))
	errOpt3 := loadgen.WithErrGaugeOpt(loadgen.DefaultErrGauge("events3_err"))

	opt1 := loadgen.WithExecLimitOpt(100000)
	opt3 := loadgen.WithExecLimitOpt(100000)
	fn1 := loadgen.NewFunction(
		[]*loadgen.Stmt{loadgen.NewStmt(streamEvents(cl, 1), 2, 5*time.Second, 5*time.Second, histOpt1, errOpt1)},
		25, opt1,
	)
	fn3 := loadgen.NewFunction(
		[]*loadgen.Stmt{loadgen.NewStmt(streamEvents(cl, 10000), 2, 5*time.Second, 5*time.Second, histOpt3, errOpt3)},
		1, opt3,
	)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	lg := loadgen.GenerateLoad(ctx, cancel, "load_test", []*loadgen.Function{fn1, fn3}, 5, 100, 5*time.Second, 1)
	fmt.Println("error:", lg.StartSever(":"+strconv.Itoa(metricsPort)))
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
