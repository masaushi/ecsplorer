package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/masaushi/ecsplorer/internal/ai"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/handler"
)

// Version is the version of `ecsplorer`, and injected at build time.
var Version = ""

// Execute executes a whole process of ecsplorer.
func Execute(args []string) {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.Usage = usage(flags)
	version := flags.Bool("version", false, "show the version of ecsplorer")
	help := flags.Bool("help", false, "help for ecsplorer")
	profile := flags.String("profile", "", "aws profile")
	cmd := flags.String("cmd", "/bin/sh", "command to execute in in the container")
	aiEnabled := flags.Bool("ai", true, "enable AI-powered analysis features")
	aiProviderName := flags.String("ai-provider", "bedrock", "AI provider (bedrock, gemini, anthropic, openai, openai-compatible)")
	aiModel := flags.String("ai-model", "", "AI model ID (provider-specific)")
	aiRegion := flags.String("ai-region", "", "AWS region for AI service (Bedrock)")
	aiAPIKey := flags.String("ai-api-key", "", "API key for AI provider (env: GEMINI_API_KEY, ANTHROPIC_API_KEY, OPENAI_API_KEY)")
	aiBaseURL := flags.String("ai-base-url", "", "custom base URL for OpenAI-compatible provider")

	if err := flags.Parse(args[1:]); err != nil {
		os.Exit(1)
	}
	if *version {
		fmt.Fprintf(os.Stdout, "ecsplorer version: %s\n", getVersion())
		os.Exit(0)
	}
	if *help {
		flags.Usage()
		os.Exit(0)
	}

	apiKey := *aiAPIKey
	if apiKey == "" {
		apiKey = aiAPIKeyFromEnv(*aiProviderName)
	}

	aiCfg := ai.Config{
		Enabled:       *aiEnabled,
		ProviderName:  *aiProviderName,
		ModelID:       *aiModel,
		BedrockRegion: *aiRegion,
		APIKey:        apiKey,
		BaseURL:       *aiBaseURL,
	}

	start, err := app.CreateApplication(context.Background(), getVersion(), *profile, cmd, aiCfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := start(handler.ClusterListHandler); err != nil {
		log.Fatal(err)
	}
}

func usage(flags *flag.FlagSet) func() {
	return func() {
		s := "ecsplorer is a tool designed for easy CLI operations with AWS ECS.\n\n" +
			"Usage of ecsplorer:\n" +
			"\tecsplorer [--flags]\n\n" +
			"for more information, see: https://github.com/masaushi/ecsplorer\n\n" +
			"flags:\n"
		fmt.Fprint(os.Stderr, s)
		flags.PrintDefaults()
	}
}

func aiAPIKeyFromEnv(provider string) string {
	switch provider {
	case "gemini":
		return os.Getenv("GEMINI_API_KEY")
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	case "openai", "openai-compatible":
		return os.Getenv("OPENAI_API_KEY")
	default:
		return ""
	}
}

func getVersion() string {
	if Version != "" {
		return "v" + Version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" {
		return "unknown"
	}

	return info.Main.Version
}
