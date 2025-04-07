// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/jumonmd/jumon/internal/client"
	"github.com/jumonmd/jumon/internal/server"
	"github.com/jumonmd/jumon/internal/version"
	"github.com/jumonmd/jumon/module"
)

// CLI definition.
var CLI struct {
	DisableTelemetry bool `help:"Disable anonymous telemetry data collection." default:"false"`
	Debug            bool `help:"Enable debug mode." default:"false"`

	Serve struct {
		Config string `help:"Config file path."`
	} `cmd:"" help:"Serve the jumon server."`

	Stop struct{} `cmd:"" help:"Stop the jumon server."`

	Init struct {
		Name string `arg:"" name:"name" help:"Name of the module."`
	} `cmd:"" help:"Initialize the JUMON.md."`

	Run struct {
		Name  string `arg:"" name:"url_or_path" help:"URL or Path to the jumon script."`
		Input string `arg:"" optional:"" name:"input" help:"Input to the module."`
	} `cmd:"" help:"Run the module."`

	Version struct{} `cmd:"" help:"Show the version."`
}

func main() {
	cli := kong.Parse(&CLI, kong.Name("jumon"), kong.Description("Magically simple markdown-based AI workflow orchestration"), kong.UsageOnError())

	if CLI.Debug {
		os.Setenv("JUMON_DEBUG", "1")
	}

	var err error
	switch cli.Command() {
	case "serve":
		err = server.Serve(CLI.DisableTelemetry)
	case "stop":
		err = server.Quit()
		if err == nil {
			log.Println("server stopped")
		}
	case "init <name>":
		err = module.InitModule(CLI.Init.Name)
	case "run <url_or_path>":
		cfg, err := client.LoadConfig(client.DefaultConfigPath())
		if err != nil {
			log.Println(err)
		}
		if err := client.WaitServer(os.Args[0], cfg.ServerURL); err != nil {
			log.Println(err)
		}
		if err := client.Run(CLI.Run.Name, []byte(CLI.Run.Input)); err != nil {
			log.Println(err)
		}
	case "version":
		fmt.Println(version.Version)
	}
	if err != nil {
		log.Println(err)
	}
}
