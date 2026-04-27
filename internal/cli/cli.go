package cli

import (
	"flag"
	"fmt"
	"io"

	"configaudit/internal/app"
	"configaudit/internal/parser"
	grpcserver "configaudit/internal/server/grpcserver"
	httpserver "configaudit/internal/server/httpserver"
)

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "server" {
		return runServerCommand(args[1:], stdout, stderr)
	}

	fs := flag.NewFlagSet("configaudit", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		useStdin         bool
		silentShort      bool
		silentLong       bool
		recursive        bool
		checkPermissions bool
		formatValue      string
		httpAddr         string
		grpcAddr         string
	)

	fs.BoolVar(&useStdin, "stdin", false, "read configuration from stdin")
	fs.BoolVar(&silentShort, "s", false, "print problems but exit with code 0")
	fs.BoolVar(&silentLong, "silent", false, "print problems but exit with code 0")
	fs.BoolVar(&recursive, "recursive", false, "scan directories recursively for .json/.yaml/.yml files")
	fs.BoolVar(&checkPermissions, "check-permissions", false, "check config file permissions with os.Stat")
	fs.StringVar(&formatValue, "format", "auto", "input format: auto, json, yaml")
	fs.StringVar(&httpAddr, "http", "", "run the HTTP server on the given address")
	fs.StringVar(&grpcAddr, "grpc", "", "run the gRPC server on the given address")

	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage:")
		fmt.Fprintln(stderr, "  configaudit [--format auto|json|yaml] [-s|--silent] <file>")
		fmt.Fprintln(stderr, "  configaudit --recursive [--check-permissions] <directory>")
		fmt.Fprintln(stderr, "  configaudit --stdin [--format auto|json|yaml] [-s|--silent] < config.yaml")
		fmt.Fprintln(stderr, "")
		fmt.Fprintln(stderr, "Server modes:")
		fmt.Fprintln(stderr, "  configaudit server --http :8080")
		fmt.Fprintln(stderr, "  configaudit server --grpc :9090")
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	format, err := parser.ParseFormat(formatValue)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}

	rest := fs.Args()
	if httpAddr != "" || grpcAddr != "" {
		if useStdin || silentShort || silentLong || recursive || checkPermissions || format != parser.FormatAuto || len(rest) > 0 {
			fmt.Fprintln(stderr, "error: scan flags and paths cannot be used together with server mode")
			fs.Usage()
			return 2
		}
		return runServers(httpAddr, grpcAddr, stdout, stderr)
	}

	if useStdin && len(rest) > 0 {
		fmt.Fprintln(stderr, "error: --stdin cannot be used together with a file path")
		fs.Usage()
		return 2
	}
	if useStdin && recursive {
		fmt.Fprintln(stderr, "error: --recursive cannot be used together with --stdin")
		fs.Usage()
		return 2
	}
	if useStdin && checkPermissions {
		fmt.Fprintln(stderr, "error: --check-permissions cannot be used together with --stdin")
		fs.Usage()
		return 2
	}
	if !useStdin && len(rest) == 0 {
		fmt.Fprintln(stderr, "error: configuration file path is required unless --stdin is used")
		fs.Usage()
		return 2
	}
	if len(rest) > 1 {
		fmt.Fprintln(stderr, "error: only one file or directory path is allowed")
		fs.Usage()
		return 2
	}

	var path string
	if len(rest) == 1 {
		path = rest[0]
	}

	return app.Run(app.RunOptions{
		Path:             path,
		UseStdin:         useStdin,
		Silent:           silentShort || silentLong,
		Recursive:        recursive,
		CheckPermissions: checkPermissions,
		Format:           format,
		Stdin:            stdin,
		Stdout:           stdout,
		Stderr:           stderr,
	})
}

func runServerCommand(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("configaudit server", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		httpAddr string
		grpcAddr string
	)

	fs.StringVar(&httpAddr, "http", "", "run the HTTP server on the given address")
	fs.StringVar(&grpcAddr, "grpc", "", "run the gRPC server on the given address")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage:")
		fmt.Fprintln(stderr, "  configaudit server --http :8080")
		fmt.Fprintln(stderr, "  configaudit server --grpc :9090")
		fmt.Fprintln(stderr, "  configaudit server --http :8080 --grpc :9090")
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(stderr, "error: server mode does not accept positional arguments")
		fs.Usage()
		return 2
	}

	return runServers(httpAddr, grpcAddr, stdout, stderr)
}

func runServers(httpAddr, grpcAddr string, stdout, stderr io.Writer) int {
	if httpAddr == "" && grpcAddr == "" {
		fmt.Fprintln(stderr, "error: at least one of --http or --grpc must be provided")
		return 2
	}

	scanner := app.NewService()
	errCh := make(chan error, 2)
	started := 0

	if httpAddr != "" {
		started++
		fmt.Fprintf(stdout, "HTTP server listening on %s\n", httpAddr)
		go func() {
			errCh <- httpserver.ListenAndServe(httpAddr, scanner)
		}()
	}
	if grpcAddr != "" {
		started++
		fmt.Fprintf(stdout, "gRPC server listening on %s\n", grpcAddr)
		go func() {
			errCh <- grpcserver.ListenAndServe(grpcAddr, scanner)
		}()
	}

	for i := 0; i < started; i++ {
		if err := <-errCh; err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 2
		}
	}

	return 0
}
