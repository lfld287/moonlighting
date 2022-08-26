/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"go.uber.org/zap/zapcore"
	"moonlighting/common/database/badgerManager"
	"moonlighting/common/logger/base"
	"moonlighting/communityServiceTradingCenter/dataManager"
	"moonlighting/communityServiceTradingCenter/httpApiServer"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "communityServiceTradingCenter",
	Short: "god bull niu",
	Long:  `god bull wu di`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.communityServiceTradingCenter.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func startServer() {
	rc := readConfig()
	l := base.NewBaseLogger(path.Join(rc.LogDir, "main.log"), 1, 1, 3, false, zapcore.InfoLevel, true)

	m := badgerManager.NewBadgerManager(l, rc.DbPath)
	go m.Start()
	defer m.Stop()

	l.Info("wait 3 second for internal db to prepare")
	<-time.After(3 * time.Second)

	providerDataManager := dataManager.NewDataManager(l, "provider.", m)
	go providerDataManager.Start()
	defer providerDataManager.Stop()

	publisherDataManager := dataManager.NewDataManager(l, "publisher.", m)
	go publisherDataManager.Start()
	defer publisherDataManager.Stop()

	recommenderDataManager := dataManager.NewDataManager(l, "recommender.", m)
	go recommenderDataManager.Start()
	defer recommenderDataManager.Stop()

	has := httpApiServer.NewHttpApiServer(rc.ServeAddress, rc.StaticServeDir, providerDataManager, publisherDataManager, recommenderDataManager)
	go has.Start()
	defer has.Stop()

	systemSignalChan := make(chan os.Signal, 1)
	signal.Notify(systemSignalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	<-systemSignalChan

}
