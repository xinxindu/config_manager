package config_manager

import (
	"fmt"

	"github.com/goroom/logger"
)

// 解析json的数据结构
type SrcDomain struct {
	DomainList []*DomainInfo `json:"domain_list"`
}

// 业务请求的URL
type UseURL struct {
	UrlPrefix string
	Params    string
}

// 业务使用的数据结构
type UseDomain struct {
	DomainInfoMap map[string]*DomainInfo
}

// 定义拉取数据的基础数据和业务使用的结构体，还需实现拉取通用配置的接口(ConfigManagerInterface)
type DomainConf struct {
	Conf   *UseDomain // 业务使用的结构体
	UseURL *UseURL    // 业务请求的URL拼装格式
}

type DomainInfo struct {
	Domain string `json:"domain"`
	IsUse  bool   `json:"is_use"`
}

// 业务读取配置的变量
var Domain *DomainConf

// 初始化Domain，并加入到confMap中
// 注意：
// 1、Name：要赋值；
// 2、Conf:要初始化业务的结构体
func init() {
	Domain = &DomainConf{
		Conf: &UseDomain{
			DomainInfoMap: make(map[string]*DomainInfo),
		},
		UseURL: &UseURL{
			UrlPrefix: "http://www.baidu.com",
			Params:    "",
		},
	}
	modelProMap.Store("domain", Domain)
}

/*
	业务需要实现的接口
*/
func (d *DomainConf) GenerateUrlFunc() string {
	if d == nil || d.UseURL == nil || d.UseURL.UrlPrefix == "" {
		return ""
	}
	if d.UseURL.Params == "" {
		return d.UseURL.UrlPrefix
	}
	return fmt.Sprintf("%s?%s", d.UseURL.UrlPrefix, d.UseURL.Params)
}

func (d *DomainConf) GetSrcObjFunc() interface{} {
	return &SrcDomain{}
}

func (d *DomainConf) SwitchObjFunc(src interface{}) error {
	obj := src.(*SrcDomain)
	target := &UseDomain{}

	tmpMap := make(map[string]*DomainInfo)
	for _, v := range obj.DomainList {
		tmpMap[v.Domain] = v
	}

	target.DomainInfoMap = tmpMap
	logger.Infof("update domain_list %v", target.DomainInfoMap)

	d.Conf = target
	return nil
}
