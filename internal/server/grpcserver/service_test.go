package grpcserver

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	configauditpb "configaudit/api/proto"
	"configaudit/internal/app"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	reflectionpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestScanReturnsFindings(t *testing.T) {
	client, cleanup := grpcTestClient(t)
	defer cleanup()

	response, err := client.Scan(context.Background(), &configauditpb.ScanRequest{
		Content: "log:\n  level: debug\n",
		Format:  "yaml",
	})
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(response.GetProblems()) != 1 {
		t.Fatalf("expected 1 problem, got %d", len(response.GetProblems()))
	}
	if response.GetProblems()[0].GetPath() != "log.level" {
		t.Fatalf("unexpected path %q", response.GetProblems()[0].GetPath())
	}
}

func TestScanRejectsInvalidConfig(t *testing.T) {
	client, cleanup := grpcTestClient(t)
	defer cleanup()

	_, err := client.Scan(context.Background(), &configauditpb.ScanRequest{
		Content: "tls:\n  enabled: [\n",
		Format:  "yaml",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %s", status.Code(err))
	}
}

func TestServerRegistersReflection(t *testing.T) {
	connection, cleanup := grpcTestConnection(t)
	defer cleanup()

	client := reflectionpb.NewServerReflectionClient(connection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.ServerReflectionInfo(ctx)
	if err != nil {
		t.Fatalf("ServerReflectionInfo returned error: %v", err)
	}
	if err := stream.Send(&reflectionpb.ServerReflectionRequest{
		MessageRequest: &reflectionpb.ServerReflectionRequest_ListServices{
			ListServices: "*",
		},
	}); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	response, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv returned error: %v", err)
	}

	if !containsService(response.GetListServicesResponse().GetService(), "configaudit.v1.ConfigAudit") {
		t.Fatalf("expected reflection services to include configaudit.v1.ConfigAudit, got %#v", response.GetListServicesResponse().GetService())
	}
}

func grpcTestClient(t *testing.T) (configauditpb.ConfigAuditClient, func()) {
	t.Helper()

	connection, cleanup := grpcTestConnection(t)
	return configauditpb.NewConfigAuditClient(connection), cleanup
}

func grpcTestConnection(t *testing.T) (*grpc.ClientConn, func()) {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	server := NewServer(app.NewService())

	go func() {
		_ = server.Serve(listener)
	}()

	dialContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	connection, err := grpc.DialContext(
		dialContext,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		cancel()
		server.Stop()
		t.Fatalf("DialContext returned error: %v", err)
	}

	cleanup := func() {
		cancel()
		if err := connection.Close(); err != nil && !strings.Contains(err.Error(), "closed") {
			t.Fatalf("Close returned error: %v", err)
		}
		server.Stop()
	}

	return connection, cleanup
}

func containsService(services []*reflectionpb.ServiceResponse, name string) bool {
	for _, service := range services {
		if service.GetName() == name {
			return true
		}
	}

	return false
}
