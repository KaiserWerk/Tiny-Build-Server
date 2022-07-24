package buildservice

import (
	"context"
	"errors"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/common"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/configuration"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/dbservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/deploymentservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/sessionstore/v2"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"net/url"
	"os/exec"
)

var (
	ErrCanceled = errors.New("buildservice: canceled by context")
)

type BuildService struct {
	Cfg       *configuration.AppConfig
	SessMgr   *sessionstore.SessionManager
	Logger    *logrus.Entry
	DBSvc     *dbservice.DBService
	DeploySvc *deploymentservice.DeploymentService
}

func New(cfg *configuration.AppConfig, sessMgr *sessionstore.SessionManager, logger *logrus.Entry,
	ds *dbservice.DBService, dpl *deploymentservice.DeploymentService) *BuildService {

	return &BuildService{
		Cfg:       cfg,
		SessMgr:   sessMgr,
		Logger:    logger.WithField("context", "buildSvc"),
		DBSvc:     ds,
		DeploySvc: dpl,
	}
}

func (bs *BuildService) CloneRepository(ctx context.Context, branch string, repositoryUrl string, path string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", "--single-branch", "--branch", branch, repositoryUrl, path)
	return cmd.Run()
}

func (bs *BuildService) GetRepositoryUrl(ctx context.Context, cont *entity.BuildDefinitionContent, withCredentials bool) (string, error) {
	if ctx.Err() != nil {
		return "", ErrCanceled
	}
	//var url string
	switch cont.Repository.Hoster {
	case "local":
		return cont.Repository.Url, nil
	default:
		urlParts, err := url.ParseRequestURI(cont.Repository.Url)
		if err != nil {
			return "", err
		}
		if !withCredentials {
			return urlParts.String(), nil
		}
		urlParts.User = url.UserPassword(cont.Repository.AccessUser, cont.Repository.AccessSecret)
		return urlParts.String(), nil
	}
}

func (bs *BuildService) GetBasePath() string {
	settings, err := bs.DBSvc.GetAllSettings()
	if err == nil {
		if path, ok := settings["base_datapath"]; ok && path != "" {
			bs.Cfg.Build.BasePath = path
		}
	}

	if bs.Cfg.Build.BasePath != "" {
		return bs.Cfg.Build.BasePath
	}

	return ""
}

func GetPreparedContent(ctx context.Context, bd *entity.BuildDefinition, vars []entity.UserVariable) (*entity.BuildDefinitionContent, error) {
	if ctx.Err() != nil {
		return nil, ErrCanceled
	}

	common.ReplaceVariables(&bd.Raw, vars)

	var bdc entity.BuildDefinitionContent
	if err := yaml.Unmarshal([]byte(bd.Raw), &bdc); err != nil {
		return nil, err
	}

	return &bdc, nil
}
