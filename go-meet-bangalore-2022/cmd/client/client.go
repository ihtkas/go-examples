package main

import (
	"context"
	"flag"
	"fmt"
	detectivepb "github.com/ihtkas/go-examples/go-meet-bangalore-2022/gen/detective"
	"github.com/ihtkas/loadgen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"strconv"
	"strings"
	"time"
)

var dCl, sdCl detectivepb.DetectiveClient

var dport int
var sdport int

var longContent = strings.Repeat(content, 1000)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.IntVar(&dport, "dport", 8123, "port to host gRPC server for detective service")
	flag.IntVar(&sdport, "sdport", 8124, "port to host gRPC server for smart detective service")
	flag.Parse()

	dconn, err := grpc.Dial("127.0.0.1:"+strconv.Itoa(dport), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	sdconn, err := grpc.Dial("127.0.0.1:"+strconv.Itoa(sdport), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	dCl = detectivepb.NewDetectiveClient(dconn)
	sdCl = detectivepb.NewDetectiveClient(sdconn)

	histOpt1 := loadgen.WithHistogramOpt(loadgen.DefaultHistogram("detective1", 0.1, 0.1, 100))
	histOpt2 := loadgen.WithHistogramOpt(loadgen.DefaultHistogram("detective2", 0.1, 0.1, 100))
	histOpt3 := loadgen.WithHistogramOpt(loadgen.DefaultHistogram("detective3", 0.1, 0.1, 100))
	histOpt4 := loadgen.WithHistogramOpt(loadgen.DefaultHistogram("detective4", 0.1, 0.1, 100))
	errOpt1 := loadgen.WithErrGaugeOpt(loadgen.DefaultErrGauge("detective1_err"))
	errOpt2 := loadgen.WithErrGaugeOpt(loadgen.DefaultErrGauge("detective2_err"))
	errOpt3 := loadgen.WithErrGaugeOpt(loadgen.DefaultErrGauge("detective3_err"))
	errOpt4 := loadgen.WithErrGaugeOpt(loadgen.DefaultErrGauge("detective4_err"))

	opt1 := loadgen.WithExecLimitOpt(100000)
	opt2 := loadgen.WithExecLimitOpt(100000)
	opt3 := loadgen.WithExecLimitOpt(1000)
	opt4 := loadgen.WithExecLimitOpt(1000)
	fn1 := loadgen.NewFunction(
		[]*loadgen.Stmt{loadgen.NewStmt(callDetectiveShort, 1, time.Second, time.Second, histOpt1, errOpt1)},
		10, opt1,
	)
	fn2 := loadgen.NewFunction(
		[]*loadgen.Stmt{loadgen.NewStmt(callSmartDetectiveShort, 1, time.Second, time.Second, histOpt2, errOpt2)},
		10, opt2,
	)
	fn3 := loadgen.NewFunction(
		[]*loadgen.Stmt{loadgen.NewStmt(callDetectiveLong, 1, time.Second, time.Second, histOpt3, errOpt3)},
		1, opt3,
	)
	fn4 := loadgen.NewFunction(
		[]*loadgen.Stmt{loadgen.NewStmt(callSmartDetectiveLong, 1, time.Second, time.Second, histOpt4, errOpt4)},
		1, opt4,
	)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	lg := loadgen.GenerateLoad(ctx, cancel, "load_test", []*loadgen.Function{fn1, fn2, fn3, fn4}, 5, 100, 5*time.Second, 1)
	fmt.Println("error:", lg.StartSever(":2112"))
}

func callDetectiveShort(ctx context.Context, iter uint, payload map[string]interface{}) (contRepeat bool, err error) {
	resp, err := dCl.FindClues(ctx, &detectivepb.FindCluesRequest{
		Content: `firstClue: "clue1" 
					secondClue: "clue2"`,
	})
	if err != nil {
		log.Println(err)
		return false, err
	}
	log.Println(len(resp.FormattedSecret))
	return true, nil
}

func callSmartDetectiveShort(ctx context.Context, iter uint, payload map[string]interface{}) (contRepeat bool, err error) {
	resp, err := sdCl.FindClues(ctx, &detectivepb.FindCluesRequest{
		Content: `firstClue: "clue1" 
					secondClue: "clue3"`,
	})
	if err != nil {
		log.Println(err)
		return false, err
	}
	log.Println(len(resp.FormattedSecret))
	return true, nil
}

func callDetectiveLong(ctx context.Context, iter uint, payload map[string]interface{}) (contRepeat bool, err error) {
	resp, err := dCl.FindClues(ctx, &detectivepb.FindCluesRequest{
		Content: longContent,
	}, grpc.MaxCallSendMsgSize(100000000))
	if err != nil {
		log.Println(err)
		return false, err
	}
	log.Println(len(resp.FormattedSecret))
	return true, nil
}

func callSmartDetectiveLong(ctx context.Context, iter uint, payload map[string]interface{}) (contRepeat bool, err error) {
	resp, err := sdCl.FindClues(ctx, &detectivepb.FindCluesRequest{
		Content: longContent,
	}, grpc.MaxCallSendMsgSize(100000000))
	if err != nil {
		log.Println(err)
		return false, err
	}
	log.Println(len(resp.FormattedSecret))
	return true, nil
}

const content = `
firstClue: "verylongclue.........................................................................................1" 
secondClue: "verylongclue.........................................................................................2"`
