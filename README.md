# config_manager
#### 通过定时读取远端或者本地的数据，来实时更新内存和本地的配置信息。
- 可以灵活的添加要读取的数据，只需实现ConfigManagerInterface接口即可
- 可以自定义请求数据的Url，数据原始结构和转化后的结构。
```go
type ConfigManagerInterface interface {
	GetSrcObjFunc() interface{}          // 获取json的结构体
	SwitchObjFunc(src interface{}) error // 转化函数
	GenerateUrlFunc() string             // 生成Url的规则
}
```
