package start

import (
	"github.com/dgmorales/go-cli-selfupdate/config"
	"github.com/dgmorales/go-cli-selfupdate/gh"
	"github.com/dgmorales/go-cli-selfupdate/kube"
	"github.com/dgmorales/go-cli-selfupdate/logger"
	"github.com/dgmorales/go-cli-selfupdate/version"
	"github.com/google/go-github/v48/github"
	"k8s.io/client-go/kubernetes"
)

type State struct {
	Version     version.Checker
	ServerCfg   config.ServerSideConfig
	Kube        *kubernetes.Clientset
	Github      *github.Client
	ssCfgLoader config.ServerSideConfigLoader
}

func ForAPIUse(debug bool) (State, error) {
	var err error

	logger.SetUp(debug)
	s := State{}

	s.Github, err = gh.NewClient()
	if err != nil {
		return State{}, err
	}

	s.Kube, err = kube.NewClient()
	if err != nil {
		return State{}, err
	}

	s.ssCfgLoader, err = config.NewKubeSSCfgLoader(s.Kube, "", "")
	if err != nil {
		return State{}, err
	}

	s.ServerCfg, err = s.ssCfgLoader.Load()
	if err != nil {
		return State{}, err
	}

	s.Version, err = version.NewGithubChecker(
		s.Github,
		s.ServerCfg.RepoOwner,
		s.ServerCfg.RepoName,
		s.ServerCfg.MinimalRequiredVersion,
		version.Current)
	if err != nil {
		return State{}, err
	}

	return s, nil
}

func ForLocalUse(debug bool) (*State, error) {
	logger.SetUp(debug)
	s := State{}
	return &s, nil
}
