package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/load"
	"github.com/tufin/oasdiff/report"
	"github.com/tufin/oasdiff/stats"
	"gopkg.in/yaml.v3"
)

var base, revision, prefix, filter, filterExtension, format string
var excludeExamples, excludeDescription, summary, breakingOnly, failOnDiff, sendStats bool

const (
	formatYAML = "yaml"
	formatText = "text"
	formatHTML = "html"
)

func init() {
	flag.StringVar(&base, "base", "", "path of original OpenAPI spec in YAML or JSON format")
	flag.StringVar(&revision, "revision", "", "path of revised OpenAPI spec in YAML or JSON format")
	flag.StringVar(&prefix, "prefix", "", "if provided, paths in base spec will be compared with 'prefix'+paths in revision spec")
	flag.StringVar(&filter, "filter", "", "if provided, diff will include only paths that match this regular expression")
	flag.StringVar(&filterExtension, "filter-extension", "", "if provided, diff will exclude paths and operations with an OpenAPI Extension matching this regular expression")
	flag.BoolVar(&excludeExamples, "exclude-examples", false, "ignore changes to examples")
	flag.BoolVar(&excludeDescription, "exclude-description", false, "ignore changes to descriptions")
	flag.BoolVar(&summary, "summary", false, "display a summary of the changes instead of the full diff")
	flag.BoolVar(&breakingOnly, "breaking-only", false, "display breaking changes only")
	flag.StringVar(&format, "format", formatYAML, "output format: yaml, text or html")
	flag.BoolVar(&failOnDiff, "fail-on-diff", false, "fail with exit code 1 if a difference is found")
	flag.BoolVar(&sendStats, "send-stats", false, "help us improve the diff tool by sending anonymous statistics")
}

func validateFlags() bool {
	if base == "" {
		fmt.Printf("please specify the '-base' flag: the path of the original OpenAPI spec in YAML or JSON format\n")
		return false
	}
	if revision == "" {
		fmt.Printf("please specify the '-revision' flag: the path of the revised OpenAPI spec in YAML or JSON format\n")
		return false
	}
	supportedFormats := map[string]bool{"": true, "yaml": true, "text": true, "html": true}
	if !supportedFormats[format] {
		fmt.Printf("invalid format. Should be yaml, text or html\n")
		return false
	}
	return true
}

const (
	statusCodeInvalidFlags    = 101
	statusCodeLoadBaseErr     = 102
	statusCodeLoadRevisionErr = 103
	statusCodeDiffErr         = 104
	statusCodeSummaryErr      = 105
	statusCodeYAMLErr         = 106
	statusCodeReportErr       = 107
	statusCodeInvalidFormat   = 108
)

func initConfig(excludeExamples bool, excludeDescription bool, filter string, filterExtension string, prefix string, breakingOnly bool) *diff.Config {
	config := diff.NewConfig()
	config.ExcludeExamples = excludeExamples
	config.ExcludeDescription = excludeDescription
	config.PathFilter = filter
	config.FilterExtension = filterExtension
	config.PathPrefix = prefix
	config.BreakingOnly = breakingOnly
	return config
}

func main() {
	times := stats.Times{}
	times.Start = time.Now()

	flag.Parse()

	config := initConfig(
		excludeExamples,
		excludeDescription,
		filter,
		filterExtension,
		prefix,
		breakingOnly,
	)

	if !validateFlags() {
		exitWithError(stats.GetInfo(statusCodeInvalidFlags, config, base, revision, times, nil, nil))
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	s1, err := load.From(loader, base)
	if err != nil {
		fmt.Printf("failed to load base spec from %q with %v\n", base, err)
		exitWithError(stats.GetInfo(statusCodeLoadBaseErr, config, base, revision, times, nil, err))
	}

	s2, err := load.From(loader, revision)
	if err != nil {
		fmt.Printf("failed to load revision spec from %q with %v\n", revision, err)
		exitWithError(stats.GetInfo(statusCodeLoadRevisionErr, config, base, revision, times, nil, err))
	}

	times.Load = time.Now()

	diffReport, err := diff.Get(config, s1, s2)
	times.Diff = time.Now()

	if err != nil {
		fmt.Printf("diff failed with %v\n", err)
		exitWithError(stats.GetInfo(statusCodeDiffErr, config, base, revision, times, nil, err))
	}

	if summary {
		if err = printYAML(diffReport.GetSummary()); err != nil {
			fmt.Printf("failed to print summary with %v\n", err)
			exitWithError(stats.GetInfo(statusCodeSummaryErr, config, base, revision, times, diffReport, err))
		}
		exitNormally(diffReport.Empty(), &stats.Info{
			Config: config,
		})
	}
	times.Summary = time.Now()

	if format == formatYAML {
		if err = printYAML(diffReport); err != nil {
			fmt.Printf("failed to print diff YAML with %v\n", err)
			exitWithError(stats.GetInfo(statusCodeYAMLErr, config, base, revision, times, diffReport, err))
		}
	} else if format == formatText {
		fmt.Printf("%s", report.GetTextReportAsString(diffReport))
	} else if format == formatHTML {
		html, err := report.GetHTMLReportAsString(diffReport)
		if err != nil {
			fmt.Printf("failed to generate HTML diff report with %v\n", err)
			exitWithError(stats.GetInfo(statusCodeReportErr, config, base, revision, times, diffReport, err))
		}
		fmt.Printf("%s", html)
	} else {
		fmt.Printf("unknown output format %q\n", format)
		exitWithError(stats.GetInfo(statusCodeInvalidFormat, config, base, revision, times, diffReport, err))
	}
	times.Output = time.Now()

	exitNormally(diffReport.Empty(), stats.GetInfo(statusCodeInvalidFormat, config, base, revision, times, diffReport, err))
}

func exitNormally(diffEmpty bool, data *stats.Info) {
	stats.Send(data)

	if failOnDiff && !diffEmpty {
		os.Exit(1)
	}
	os.Exit(0)
}

func exitWithError(data *stats.Info) {
	stats.Send(data)
	os.Exit(data.StatusCode)
}

func printYAML(output interface{}) error {
	if reflect.ValueOf(output).IsNil() {
		return nil
	}

	bytes, err := yaml.Marshal(output)
	if err != nil {
		return err
	}
	fmt.Printf("%s", bytes)
	return nil
}
