package smartdetectiveservice

import (
	"context"
	detectivepb "github.com/ihtkas/go-examples/go-meet-bangalore-2022/gen/detective"
	"regexp"
	"strings"
)

type DetectiveService struct {
	detectivepb.UnimplementedDetectiveServer
}

var cluesRegex = regexp.MustCompile(`Clue: "(.*)"`)

func (s *DetectiveService) FindClues(ctx context.Context, req *detectivepb.FindCluesRequest) (*detectivepb.FindCluesResponse, error) {
	matches := cluesRegex.FindAllStringSubmatch(req.Content, -1)
	res := &strings.Builder{}
	for _, matches := range matches {
		_, err := res.WriteString(matches[1])
		if err != nil {
			return nil, err
		}
		_, err = res.WriteString("\n")
		if err != nil {
			return nil, err
		}
	}
	return &detectivepb.FindCluesResponse{
		FormattedSecret: res.String(),
	}, nil
}
