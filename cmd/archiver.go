package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/TimH8/gotrain/archiver"

	"github.com/TimH8/gotrain/receiver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var archiverCommand = &cobra.Command{
	Use:   "archiver",
	Short: "Start archiver",
	Long:  `Start the GoTrain archiver. It receives data and pushes processed data to the archive queue.`,
	Run: func(cmd *cobra.Command, args []string) {
		startArchiver(cmd)
	},
}

func init() {
	RootCmd.AddCommand(archiverCommand)
}

var exitArchiverReceiverChannel = make(chan bool)

func startArchiver(cmd *cobra.Command) {
	initLogger(cmd)

	log.Infof("GoTrain archiver %v starting", Version.VersionStringLong())

	signalChan := make(chan os.Signal, 1)
	shutdownArchiverFinished := make(chan struct{})

	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		log.Errorf("Received signal: %+v, shutting down", sig)
		signal.Reset()
		shutdownArchiver()
		close(shutdownArchiverFinished)
	}()

	connectionError := archiver.Connect()

	if connectionError != nil {
		log.WithError(connectionError).Error("Error while connecting to archive queue")

		return
	}

	receiver.ProcessStores = false
	receiver.ArchiveServices = true

	go receiver.ReceiveData(exitArchiverReceiverChannel)

	<-shutdownArchiverFinished
	log.Error("Exiting")
}

func shutdownArchiver() {
	log.Warn("Shutting down")

	exitArchiverReceiverChannel <- true

	<-exitArchiverReceiverChannel
}
