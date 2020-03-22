package config_manager

type BaseInfo struct {
	Name       string      // 通用配置的模块名
	NeedChange *ChangeInfo // 每次都需要更新的
}

type ChangeInfo struct {
	data       []byte // 存放拉取的数据
	dataMD5    string // MD5校验值
	lastReqUrl string // 最近一次请求的Url
}

type ModelData struct {
	BaseInfo *BaseInfo              // 基础数据的结构
	ModelI   ConfigManagerInterface // 业务实现的配置接口
}

type ConfigManagerInterface interface {
	GetSrcObjFunc() interface{}          // 获取json的结构体
	SwitchObjFunc(src interface{}) error // 转化函数
	GenerateUrlFunc() string             // 生成Url的规则
}

type ConsoleResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List []struct {
			ID          string `json:"id"`
			ConfigName  string `json:"config_name"`
			Param       string `json:"param"`
			Data        string `json:"data"`
			Description string `json:"description"`
		} `json:"list"`
	} `json:"data"`
}

func (c *ChangeInfo) GetLastReqUrl() string {
	if c == nil {
		return ""
	}
	return c.lastReqUrl
}

func (c *ChangeInfo) GetDataMD5() string {
	if c == nil {
		return ""
	}
	return c.dataMD5
}

func (c *ChangeInfo) GetData() []byte {
	if c == nil {
		return nil
	}
	return c.data
}

func (c *ChangeInfo) SetLastReqUrl(url string) {
	if c == nil {
		return
	}
	c.lastReqUrl = url
}
func (c *ChangeInfo) SetDataMD5(md5 string) {
	if c == nil {
		return
	}
	c.dataMD5 = md5
}
func (c *ChangeInfo) SetData(data []byte) {
	if c == nil {
		return
	}
	c.data = data
}
