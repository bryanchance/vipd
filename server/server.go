package server

import (
	"time"

	"github.com/ehazlett/element"
	"github.com/ehazlett/vipd"
	"github.com/sirupsen/logrus"
)

const (
	labelPeerStarted = "vipd.started"
)

type Server struct {
	config   *vipd.Config
	agent    *element.Agent
	started  time.Time
	updateCh chan *element.NodeEvent
}

func NewServer(cfg *vipd.Config) (*Server, error) {
	peer := &element.Peer{
		ID:      cfg.NodeName,
		Address: cfg.AdvertiseAddress,
		Labels: map[string]string{
			labelPeerStarted: time.Now().UTC().String(),
		},
	}
	agent, err := element.NewAgent(peer, &element.Config{
		ConnectionType:         "lan",
		ClusterAddress:         cfg.ClusterAddress,
		AdvertiseAddress:       cfg.AdvertiseAddress,
		LeaderPromotionTimeout: cfg.LeaderPromotionTimeout.Duration,
		Peers:                  cfg.Peers,
	})
	if err != nil {
		return nil, err
	}

	updateCh := agent.Subscribe()

	return &Server{
		config:   cfg,
		agent:    agent,
		started:  time.Now(),
		updateCh: updateCh,
	}, nil
}

func (s *Server) Run() error {
	if err := s.validateVIPs(); err != nil {
		return err
	}
	if err := s.agent.Start(); err != nil {
		return err
	}

	go s.nodeEventHandler()

	leader, err := s.getLeaderPeerID()
	if err != nil {
		logrus.WithError(err).Warn("unable to get current leader")
	}

	logrus.Debugf("leader: %s", leader)
	if err := s.reconcile(leader); err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	s.agent.Unsubscribe(s.updateCh)
	// remove vips
	if err := s.removeVIPs(); err != nil {
		logrus.WithError(err).Error("error removing VIPs")
	}
	return s.agent.Shutdown()
}
