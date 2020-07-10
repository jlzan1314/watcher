package cmd

import (
	"fmt"
	"os"
	"watcher/lib"

	"github.com/spf13/cobra"
)

var (
	Cmd   string
	Match string
	T     int
	Args  []string
	Dirs  []string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "watch",
	Short: "watch",
	Long:  `watch`,
	Run: func(c *cobra.Command, a []string) {
		c.MarkFlagRequired("cmd")
		argsObject := &lib.Args{
			Cmd:   Cmd,
			T:     T,
			Args:  Args,
			Dirs:  Dirs,
			Match: Match,
		}

		lib.Watch(argsObject)
	},
}

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "watch test",
	Long:  `watch test`,
	Run: func(c *cobra.Command, a []string) {
		fmt.Print("test cmd args:")
		fmt.Println(a)
		for {

		}
	},
}

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "watch run",
	Long:  `watch run`,
	Run: func(c *cobra.Command, a []string) {
		c.MarkFlagRequired("cmd")
		argsObject := lib.Args{
			Cmd:   Cmd,
			Args:  Args,
			Dirs:  Dirs,
			Match: Match,
		}
		lib.StartProcess(argsObject)
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.AddCommand(TestCmd)
	RootCmd.AddCommand(RunCmd)
	RootCmd.Flags().StringVarP(&Cmd, "cmd", "c", "", "cmd")
	RootCmd.Flags().StringVarP(&Match, "match", "m", "", "match 'php|js'")
	RootCmd.Flags().StringSliceVarP(&Args, "args", "a", nil, "-a args2 -a arg2")
	RootCmd.Flags().StringSliceVarP(&Dirs, "dirs", "d", nil, "-d args2 -d arg2")
	RootCmd.Flags().IntVarP(&T, "time", "t", 2, "-t 2")

	RunCmd.Flags().StringVarP(&Cmd, "cmd", "c", "", "cmd")
	RunCmd.Flags().StringVarP(&Match, "match", "m", "", "match 'php|js'")
	RunCmd.Flags().StringSliceVarP(&Args, "args", "a", nil, "-a args2 -a arg2")
	RunCmd.Flags().StringSliceVarP(&Dirs, "dirs", "d", nil, "-d args2 -d arg2")

}
