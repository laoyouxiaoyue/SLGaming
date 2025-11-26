// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Nacos NacosConf `json:",optional"`
}

type NacosConf struct {
	Hosts     []string `json:",optional"`
	Namespace string   `json:",optional"`
	Group     string   `json:",optional"`
	DataId    string   `json:",optional"`
	Username  string   `json:",optional"`
	Password  string   `json:",optional"`
}
