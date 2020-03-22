package main

import (
	"github.com/goroom/logger"
	"github.com/goroom/utils/conv"
	"github.com/xinxindu/config_manager"
)

func main() {
	config_manager.Init(
		config_manager.WithRemote(false),
		config_manager.WithConfigNameList([]string{"domain"}))
	config_manager.Wait()

	logger.Infof("domain use: %s", conv.ToJsonString(config_manager.Domain.Conf))
	logger.Infof("conf req:%v", conv.ToJsonString(config_manager.GetLastRequestURLMap()))
	select {}
}
