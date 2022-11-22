package smarteventsservice

import (
	eventspb "github.com/ihtkas/go-examples/go-meet-bangalore-2022/gen/events"
	"io"
	"log"
	"regexp"
)

type EventsService struct {
	eventspb.UnimplementedEventsServiceServer
}

var cluesRegex = regexp.MustCompile(`Clue: "(.*)"`)

func (s *EventsService) PublishEvent(server eventspb.EventsService_PublishEventServer) error {
	for {
		ctx := server.Context()
		select {
		case <-ctx.Done():
			return nil
		default:
			req, err := server.Recv()
			if err == io.EOF {
				sendErr := server.Send(&eventspb.PublishEventResponse{})
				if sendErr != nil {
					log.Println(sendErr)
				}
				return nil
			}
			log.Println("received event req...", len(req.Event))
		}
	}
}
