package cmd

import (
    "context"
    "fmt"
    "os"

    "github.com/google/go-github/v39/github"
    "github.com/spf13/cobra"
    "golang.org/x/oauth2"

    "ghtui/ghtui/ui"
)

var Username string
var Token string

var rootCmd = &cobra.Command{
    Use:     "ghtui",
    Short:   "A terminal UI for GitHub.",
    Long:    "ghtui allows you to browse and interact with GitHub from your terminal.",
    Example: "ghtui --token <token> --username <username>",
    Run: func(cmd *cobra.Command, args []string) {
        Username = getVariable(cmd, "GitHub username", "username", "GITHUB_USERNAME")
        Token = getVariable(cmd, "GitHub access token", "token", "GITHUB_TOKEN")
        ctx := context.Background()
        ts := oauth2.StaticTokenSource(
            &oauth2.Token{AccessToken: Token},
        )
        tc := oauth2.NewClient(ctx, ts)
        gh := github.NewClient(tc)
        if err := ui.NewProgram(Username, gh).Start(); err != nil {
            fmt.Println("Could not start ghtui", err)
            os.Exit(1)
        }
    },
}

func init() {
    rootCmd.PersistentFlags().StringVarP(&Username, "username", "u", "", "GitHub username")
    rootCmd.PersistentFlags().StringVarP(&Token, "token", "t", "", "GitHub personal access token")
}

func Execute() error {
    return rootCmd.Execute()
}

func getVariable(cmd *cobra.Command, name string, param string, env string) string {
    value, err := cmd.PersistentFlags().GetString(param)
    if err != nil {
        fmt.Println("Could not get "+param, err)
        os.Exit(1)
    }

    if value != "" {
        return value
    }

    if os.Getenv(env) != "" {
        return os.Getenv(env)
    }

    fmt.Println("You must pass your " + name + " using the --" + param + " flag or set the " + env + " environment variable.")
    os.Exit(1)
    return ""
}
