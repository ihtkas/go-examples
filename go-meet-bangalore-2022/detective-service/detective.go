package detectiveservice

import (
	"context"
	detectivepb "github.com/ihtkas/go-examples/go-meet-bangalore-2022/gen/detective"
	"regexp"
)

type DetectiveService struct {
	detectivepb.UnimplementedDetectiveServer
}

var cluesRegex = regexp.MustCompile(`Clue: "(.*)"`)

func (s *DetectiveService) FindClues(ctx context.Context, req *detectivepb.FindCluesRequest) (*detectivepb.FindCluesResponse, error) {
	matches := cluesRegex.FindAllStringSubmatch(req.Content, -1)
	res := ""
	for _, matches := range matches {
		res += matches[1] + "\n"
	}
	return &detectivepb.FindCluesResponse{
		FormattedSecret: res,
	}, nil
}
