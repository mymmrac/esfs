//go:generate go build -o ../../bin/esfs github.com/mymmrac/esfs/cmd/esfs

package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/mymmrac/esfs"
)

var rootCmd = &cobra.Command{
	Use:     "esfs",
	Short:   "ESFS",
	Long:    "Embeddable Serve for File System\nGitHub: https://github.com/mymmrac/esfs",
	Version: version(),
	RunE: func(cmd *cobra.Command, args []string) error {
		ops := []esfs.Option{
			esfs.WithGracefulShutdown(),
		}

		if pathRewriteToRoot {
			ops = append(ops, esfs.WithPathRewriteToRoot())
		}

		return esfs.ServeDir(addr, dir, ops...)
	},
}

var (
	addr string
	dir  string

	pathRewriteToRoot bool
)

func main() {
	rootCmd.Flags().StringVarP(&addr, "addr", "a", "", "Address of the server")
	err := rootCmd.MarkFlagRequired("addr")
	assert(err == nil, err)

	rootCmd.Flags().StringVarP(&dir, "dir", "d", "", "Directory to serve")
	err = rootCmd.MarkFlagRequired("dir")
	assert(err == nil, err)
	err = rootCmd.MarkFlagDirname("dir")
	assert(err == nil, err)

	rootCmd.Flags().BoolVarP(&pathRewriteToRoot, "fallback-to-root", "f", false,
		"Rewrite path to root if not found")

	if err = rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func version() string {
	build, ok := debug.ReadBuildInfo()
	assert(ok, "no build info found")

	var (
		err       error
		commit    string
		buildTime time.Time
		modified  bool
	)

	for _, setting := range build.Settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
		case "vcs.time":
			buildTime, err = time.Parse(time.RFC3339, setting.Value)
			assert(err == nil, fmt.Errorf("parse build time: %w", err))
		case "vcs.modified":
			modified, err = strconv.ParseBool(setting.Value)
			assert(err == nil, fmt.Errorf("parse modifed: %w", err))
		}
	}

	return fmt.Sprintf("commit: %s (modified :%t), build time: %s", commit, modified, buildTime.Local())
}

func assert(ok bool, args ...any) {
	if !ok {
		fmt.Println(append([]any{"FATAL:"}, args...)...)
		os.Exit(1)
	}
}
