package config_manager

import (
	"time"

	"github.com/goroom/utils/path"
)

type Options struct {
	configNameList []string      // 需要读取的配置列表
	t              time.Time     // 程序启动时间
	localDir       string        // 默认是当前路径
	isRemote       bool          // 默认是true
	interval       time.Duration // 默认是5s
}

func newDefaultOptions() *Options {
	return &Options{
		t:        time.Now(),
		localDir: path.GetCurrentDirectory(),
		isRemote: true,
		interval: 5,
	}
}

func (o *Options) Init(opts []Option) {
	if o == nil {
		return
	}

	for _, opt := range opts {
		opt(o)
	}
}

type Option func(*Options)

// WithRemote 是否加载远程配置
// 默认true
// 如果设置成false，优先从本地加载，并定时拉取本地数据更新缓存; 本地加载失败时从远端拉取一次存入本地，定时从本地加载刷新缓存
func WithRemote(isRemote bool) Option {
	return func(o *Options) {
		o.isRemote = isRemote
	}
}

// WithConfigNameList 需要加载的配置名列表
// 如果这里不设置，后续获取配置的时候不会自动加载
func WithConfigNameList(configNameList []string) Option {
	return func(o *Options) {
		o.configNameList = configNameList
	}
}

// WithLocalDataDir 设置本地存储文件的路径
// 只支持可执行程序所在目录的相对路径(并非启动程序时的当前目录)
// 默认 .   --> 暂不使用
func WithLocalDataDir(dir string) Option {
	return func(o *Options) {
		o.localDir = path.GetCurrentDirectory() + "/" + dir
	}
}
