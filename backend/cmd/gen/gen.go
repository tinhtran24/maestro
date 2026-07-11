// Command gen is the Thanos code-generation entrypoint. It
// dispatches on a subcommand so more generators can be added later:
//
//	gen spec [-out openapi.yaml]   # write the code-first OpenAPI document
//
// The `spec` generator is invoked via `go generate` (see
// internal/httpd/apispec/gen.go); its output openapi.yaml is committed and
// embedded by the apispec package.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tinhtran/thanos/backend/internal/httpd/apispec/specgen"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	sub := os.Args[1]
	args := os.Args[2:]

	switch sub {
	case "spec":
		runSpec(args)
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "gen: unknown subcommand %q\n\n", sub)
		usage()
		os.Exit(2)
	}
}

// runSpec writes the code-first OpenAPI document produced by specgen.Build().
func runSpec(args []string) {
	fs := flag.NewFlagSet("spec", flag.ExitOnError)
	out := fs.String("out", "openapi.yaml", "output path for the generated OpenAPI document")
	_ = fs.Parse(args)

	doc, err := specgen.Build()
	if err != nil {
		log.Fatalf("gen spec: build openapi: %v", err)
	}
	if err := os.WriteFile(*out, doc, 0o600); err != nil {
		log.Fatalf("gen spec: write %s: %v", *out, err)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `gen — Thanos code generators

Usage:
  gen <subcommand> [flags]

Subcommands:
  spec    Write the code-first OpenAPI document (flags: -out <path>)
`)
}
