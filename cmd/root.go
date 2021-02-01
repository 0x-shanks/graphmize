/*
Copyright © 2020 hourglasshoro <hourglasshoro.628@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/hourglasshoro/graphmize/pkg/file"
	"github.com/hourglasshoro/graphmize/pkg/graph"
	"github.com/hourglasshoro/graphmize/pkg/graph_path"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "graphmize",
	Short: "Graphmize is a tool to visualize the dependencies of kustomize",
	Long: `
Graphmize is a tool to visualize the dependencies of kustomize.
You can open a dashboard in your browser and see a graph of dependencies represented as a directed graph.
`,
	Version: "v0.1.1",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		source := cmd.Flag("source").Value.String()
		defaultFileSystem := afero.NewOsFs()
		ctx := file.NewContext(defaultFileSystem)
		currentDir, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "cannot get current dir")
		}
		graphDir := graph_path.Solve(source, currentDir)
		graph, err := graph.BuildGraph(*ctx, graphDir)
		if err != nil {
			return errors.Wrap(err, "cannot build graph")
		}
		graph.ToTree()
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.graphmize.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.Flags().StringP("source", "s", "", "Directory to search")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".graphmize" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".graphmize")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
