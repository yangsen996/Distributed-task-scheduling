package worker

import (
	"crontab/utils"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var G_config *Config

type Config struct {
	EtcdEndPoints         []string `json:"etcdEndPoints"`
	EtcdDialTimeout       int      `json:"etcdDialTimeout"`
	MongodbUri            string   `json:"mongodbUri"`
	MongodbConnectTimeout int      `json:"mongodbConnectTimeout"`
}

var ConfFile string

func InitConfig() (err error) {

	// // 读取文件
	// content, err := ioutil.ReadFile(filename)
	// if err != nil {
	// 	return
	// }
	// var cfg Config
	// // 反序列化
	// err = json.Unmarshal(content, &cfg)
	// if err != nil {
	// 	return
	// }
	// G_config = &cfg
	// return
	//拼接
	str, _ := os.Getwd()
	var bulid strings.Builder
	bulid.Write([]byte(str))
	bulid.WriteString("\\worker\\main\\worker.json")
	cfg := bulid.String()
	// 文件是否存在
	if !utils.Exist(cfg) {
		fmt.Println("配置文件未找到，请创建文件")
		return
	}
	ConfFile = cfg
	content, err := utils.GetConfigInfo(cfg)
	if err != nil {
		fmt.Println("文件")
	}
	var c Config
	err = json.Unmarshal([]byte(content), &c)
	if err != nil {
		fmt.Println("解析文件", cfg, "失败", err)
	}
	G_config = &c
	fmt.Println(c)
	return
}
