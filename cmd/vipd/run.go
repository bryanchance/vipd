package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/ehazlett/vipd"
	"github.com/ehazlett/vipd/server"
	"github.com/ehazlett/vipd/version"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

func runAction(clix *cli.Context) error {
	logrus.Infof("starting %s", version.FullVersion())
	var cfg *vipd.Config
	if _, err := toml.DecodeFile(clix.String("config"), &cfg); err != nil {
		if os.IsNotExist(err) {
			return errors.New("no config file specified")
		}
		return err
	}

	if cfg.NodeName == "" {
		cfg.NodeName = getHostname()
	}
	if len(cfg.VirtualIPs) == 0 {
		return fmt.Errorf("at least one virtual IP must be specified in the config")
	}
	if cfg.ClusterAddress == "" {
		return fmt.Errorf("ClusterAddress must be specified in the config")
	}
	if cfg.AdvertiseAddress == "" {
		cfg.AdvertiseAddress = cfg.ClusterAddress
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		return err
	}
	if err := srv.Run(); err != nil {
		return err
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	doneCh := make(chan bool, 1)

	go func() {
		for {
			select {
			case sig := <-signals:
				switch sig {
				case syscall.SIGTERM, syscall.SIGINT:
					logrus.Info("shutting down")
					if err := srv.Stop(); err != nil {
						logrus.Error(err)
					}
					doneCh <- true
				default:
					logrus.Warnf("unhandled signal %s", sig)
				}
			}
		}
	}()

	<-doneCh

	return nil
}
