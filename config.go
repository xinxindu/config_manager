package config_manager

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/goroom/logger"
	"github.com/goroom/utils/conv"
)

type ConfigManager struct {
	ConfMap *sync.Map // 模块数据的缓存  string -> *ModelData
}

var (
	opts          *Options
	configManager *ConfigManager
	modelProMap   = sync.Map{} // 存放各配置业务数据  string -> *ConfigManagerInterface
)

func Init(ops ...Option) {
	opts = newDefaultOptions()
	opts.Init(ops)

	configManager = &ConfigManager{
		ConfMap: updateToModelDataMap(),
	}
	configManager.start()
}

func updateToModelDataMap() *sync.Map {
	modeDataMap := &sync.Map{}
	modelProMap.Range(func(key, value interface{}) bool {
		modeDataMap.Store(key.(string), &ModelData{
			BaseInfo: &BaseInfo{Name: key.(string),
				NeedChange: &ChangeInfo{},
			},
			ModelI: value.(ConfigManagerInterface),
		})
		return true
	})
	return modeDataMap
}

func GetInstance() *ConfigManager {
	return configManager
}

func (c *ConfigManager) start() {
	for _, confName := range opts.configNameList {
		conf, ok := c.ConfMap.Load(confName)
		if !ok {
			logger.Errorf("config name is not exists in common")
			continue
		}
		go c.updateConf(confName, conf)
	}
}

func Wait() {
	for {
		ready := true
		for _, confName := range opts.configNameList {
			conf, ok := configManager.ConfMap.Load(confName)
			if !ok {
				continue
			}
			if conf.(*ModelData).BaseInfo.NeedChange.GetDataMD5() == "" {
				ready = false
				break
			}
		}
		if ready {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	logger.Infof("ConfigManager wait load [%s] config use time %dms", conv.ToJsonString(opts.configNameList), time.Since(opts.t)/time.Millisecond)
}

// GetLastRequestURLMap 获取最后一次发起请求的url
// 返回 map[config_name]url
func GetLastRequestURLMap() map[string]string {
	m := map[string]string{}
	for _, name := range opts.configNameList {
		v, ok := configManager.ConfMap.Load(name)
		if !ok {
			continue
		}
		m[name] = v.(*ModelData).BaseInfo.NeedChange.GetLastReqUrl()
	}
	return m
}

func (c *ConfigManager) loadRemoteData(modelName, lastUrl string) ([]byte, error) {
	res, err := http.Get(lastUrl)
	if err != nil {
		return nil, fmt.Errorf("http get from %v error, %v", lastUrl, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res == nil || res.StatusCode != 200 {
		var code int
		if res != nil {
			code = res.StatusCode
		}
		return nil, fmt.Errorf("http get from %v error, %v [%d]", lastUrl, err, code)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read http response data from %s error, %v", lastUrl, err)
	}

	cfg := &ConsoleResponse{}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal http response data form %s error, %v, data[%s]", lastUrl, err, string(data))
	}

	if cfg.Code != 0 || len(cfg.Data.List) == 0 {
		return nil, fmt.Errorf("error result code %v len %d, %s, %s", cfg.Code, len(cfg.Data.List), lastUrl, string(data))
	}
	if cfg.Data.List[0].ConfigName != modelName {
		return nil, fmt.Errorf("error config name %s, need %s, %s, %s", cfg.Data.List[0].ConfigName, modelName, lastUrl, string(data))
	}

	return []byte(cfg.Data.List[0].Data), nil
}

func (c *ConfigManager) loadLocalFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

func (c *ConfigManager) loadLocal(confName string, model *ModelData) error {
	data, err := c.loadLocalFile(c.getFilePath(confName))
	if err != nil {
		logger.Errorf("load local conf file[%s] err:%v", c.getFilePath(confName), err)
		return err
	}
	return c.updateAndSave(data, model, confName, false)
}

func (c *ConfigManager) loadRemote(confName string, modelData *ModelData) error {
	data, err := c.loadRemoteData(confName, modelData.BaseInfo.NeedChange.GetLastReqUrl())
	if err != nil {
		return err
	}
	return c.updateAndSave(data, modelData, confName, true)
}

func (c *ConfigManager) updateAndSave(data []byte, modelData *ModelData, confName string, save bool) error {
	// 判等MD5
	md5sum := fmt.Sprintf("%x", md5.Sum(data))
	if modelData.BaseInfo.NeedChange.GetDataMD5() == md5sum {
		return nil
	}

	// 解析原数据
	srcObj := modelData.ModelI.GetSrcObjFunc()
	err := json.Unmarshal(data, srcObj)
	if err != nil {
		logger.Errorf("json unmarshal base[%s] data err:%v", confName, err)
		return err
	}

	// 转化成新的结构
	err = modelData.ModelI.SwitchObjFunc(srcObj)
	if err != nil {
		logger.Errorf("switch struct for base[%s] err:%v", confName, err)
		return err
	}

	// 存盘
	if save {
		c.save(confName, data)
	}

	// 设置新数据
	modelData.BaseInfo.NeedChange.dataMD5 = md5sum
	modelData.BaseInfo.NeedChange.data = data
	return nil
}

func (c *ConfigManager) getFilePath(confName string) string {
	return opts.localDir + "/" + confName + ".json"
}

// 持续更新模块的配置
func (c *ConfigManager) updateConf(modelName string, modelDataI interface{}) {
	data := modelDataI.(*ModelData)
	model := data.ModelI
	var err error
	for {
		// 生成并更新请求的url
		lastUrl := model.GenerateUrlFunc()
		data.BaseInfo.NeedChange.SetLastReqUrl(lastUrl)

		if opts.isRemote {
			err = c.loadRemote(modelName, data)
		} else {
			err = c.loadLocal(modelName, data)
			if err != nil {
				logger.Errorf("[ConfigManager] load %s from %s error, try to load from remote, %v",
					modelName, c.getFilePath(modelName), err)
				err = c.loadRemote(modelName, data)
			}
		}

		if err != nil {
			logger.Errorf("[ConfigManager] load %s error, %v", modelName, err)
			time.Sleep(time.Second)
		}

		time.Sleep(opts.interval * time.Second)
	}
}

func (c *ConfigManager) save(confName string, data []byte) {
	err := os.MkdirAll(opts.localDir, os.ModePerm)
	if err != nil {
		logger.Errorf("[ConfigManager] %v", err)
		return
	}

	err = ioutil.WriteFile(c.getFilePath(confName), data, os.ModePerm)
	if err != nil {
		logger.Errorf("[ConfigManager] write config[%s] to %s error, %v", confName, c.getFilePath(confName), err)
	}
}
