package server

import (
	"time"

	"github.com/ehazlett/element"
	"github.com/sirupsen/logrus"
)

func (s *Server) nodeEventHandler() {
	for evt := range s.updateCh {
		switch evt.Type {
		case element.NodeLeave:
			logrus.Debugf("node leave event: %+v", evt)
			newLeader, err := s.getLeaderPeerID()
			if err != nil {
				logrus.WithError(err).Error("error getting leader")
				continue
			}

			logrus.Debugf("leader: %s", newLeader)
			if err := s.reconcile(newLeader); err != nil {
				logrus.WithError(err).Error("error reconciling vips")
			}
		}
	}
}

func (s *Server) getLeaderPeerID() (string, error) {
	peers, err := s.agent.Peers()
	if err != nil {
		return "", err
	}

	leader := ""
	leaderStarted := time.Now()
	for _, peer := range peers {
		format := "2006-01-02 15:04:05 -0700 MST"
		peerStarted, ok := peer.Labels[labelPeerStarted]
		if !ok {
			logrus.Warnf("peer %s does not have started label", peer.ID)
			continue
		}
		logrus.Debugf("checking peer %s: %s", peer.ID, peerStarted)
		started, err := time.Parse(format, peerStarted)
		if err != nil {
			logrus.WithError(err).Errorf("error parsing peer %s start time", peer.ID)
			continue
		}
		if started.Before(leaderStarted) {
			leader = peer.ID
			leaderStarted = started
		}
	}

	return leader, nil
}

func (s *Server) reconcile(leader string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if leader == s.currentLeader {
		logrus.Debugf("no leader change; skipping reconcile")
		return nil
	}

	s.currentLeader = leader

	if leader == s.config.NodeName {
		return s.updateVIPs()
	}
	return s.removeVIPs()
}
