package grpcserver

import (
	"context"
	"net"

	configauditpb "configaudit/api/proto"
	"configaudit/internal/app"
	"configaudit/internal/parser"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Service struct {
	configauditpb.UnimplementedConfigAuditServer
	scanner app.Scanner
}

func NewService(scanner app.Scanner) *Service {
	return &Service{scanner: scanner}
}

func (s *Service) Scan(ctx context.Context, req *configauditpb.ScanRequest) (*configauditpb.ScanResponse, error) {
	format, err := parser.ParseFormat(req.GetFormat())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid format: %v", err)
	}

	problems, err := s.scanner.ScanContent("", []byte(req.GetContent()), format)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "scan request: %v", err)
	}

	response := &configauditpb.ScanResponse{
		Problems: make([]*configauditpb.Problem, 0, len(problems)),
	}
	for _, problem := range problems {
		response.Problems = append(response.Problems, &configauditpb.Problem{
			Severity:       string(problem.Severity),
			Path:           problem.Path,
			Message:        problem.Message,
			Recommendation: problem.Recommendation,
		})
	}

	return response, nil
}

func ListenAndServe(addr string, scanner app.Scanner) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	server := NewServer(scanner)

	return server.Serve(listener)
}

func NewServer(scanner app.Scanner) *grpc.Server {
	server := grpc.NewServer()
	configauditpb.RegisterConfigAuditServer(server, NewService(scanner))
	reflection.Register(server)

	return server
}
