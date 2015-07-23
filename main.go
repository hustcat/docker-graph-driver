package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/hustcat/docker-graph-driver/api"
	"github.com/hustcat/docker-graph-driver/driver"
	"os"
)

const (
	socketAddress = "/run/docker/plugins/rbd.sock"
)

var (
	root         string
	graphDriver  string
	graphOptions []string
	flDebug      bool
	flLogLevel   string
)

func init() {
	installFlags()
}

func installFlags() {
	flag.BoolVar(&flDebug, []string{"D", "-debug"}, false, "Enable debug mode")
	flag.StringVar(&flLogLevel, []string{"l", "-log-level"}, "info", "Set the logging level")
	flag.StringVar(&root, []string{"g", "-graph"}, "/var/lib/docker", "Path to use as the root of the graph driver")
	opts.ListVar(&graphOptions, []string{"-storage-opt"}, "Set storage driver options")
	flag.StringVar(&graphDriver, []string{"s", "-storage-driver"}, "", "Force the runtime to use a specific storage driver")
}

func main() {

	flag.Parse()

	if flLogLevel != "" {
		lvl, err := logrus.ParseLevel(flLogLevel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse logging level: %s\n", flLogLevel)
			os.Exit(1)
		}
		logrus.SetLevel(lvl)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if flDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	driver, err := graphdriver.New(root, graphOptions)
	if err != nil {
		logrus.Errorf("Create rbd driver failed: %v", err)
		os.Exit(1)
	}
	h := api.NewHandler(driver)
	logrus.Infof("listening on %s\n", socketAddress)
	fmt.Println(h.ServeUnix("root", socketAddress))
}
