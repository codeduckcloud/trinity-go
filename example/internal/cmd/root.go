package cmd

import (
	"fmt"
	"log"
	"os"
	"trinity-example/internal/consts"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   consts.ProjectName,
	Short: fmt.Sprintf("%v command line tool", consts.ProjectName),
	Long:  fmt.Sprintf("%v command line tool, generated by trinity-go ", consts.ProjectName),
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("%v version: %v", consts.ProjectName, consts.Version)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("root cmd execute failed, error:%v", err)
		os.Exit(1)
	}
}