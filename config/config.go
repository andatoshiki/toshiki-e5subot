package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

const (
	LogBasePath    string = "./log/"
	WelcomeContent string = "欢迎光临俊樹のE5subot! ヾ(≧▽≦*)o"
	HelpContent    string = `
	以下是常用命令哦~
	/my 查看已绑定账户信息
	/bind  绑定新账户
	/unbind 解绑账户
	/export 导出账户信息(JSON)
	/help 帮助 (你是笨蛋嘛)
	/task 管理员手动调用一次API
	/log 管理员获取机器人历史日志
	开源地址：github.com/andatoshiki/toshiki-e5subot
	使用教程: https://telegra.ph/%E4%BF%8A%E6%A8%B9%E3%81%AEE5subot%E4%BD%BF%E7%94%A8%E6%95%99%E7%A8%8B-08-02
`
)

func Init() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		zap.S().Fatalw("failed to read config", "error", err)
	}
	BotToken = viper.GetString("bot_token")
	Cron = viper.GetString("cron")
	Socks5 = viper.GetString("socks5")

	viper.SetDefault("errlimit", 5)
	viper.SetDefault("bindmax", 5)
	viper.SetDefault("goroutine", 10)

	BindMaxNum = viper.GetInt("bindmax")
	MaxErrTimes = viper.GetInt("errlimit")
	Notice = viper.GetString("notice")

	MaxGoroutines = viper.GetInt("goroutine")
	Admins = getAdmins()
	DB = viper.GetString("db")
	Table = viper.GetString("table")

	switch DB {
	case "mysql":
		Mysql = mysqlConfig{
			Host:     viper.GetString("mysql.host"),
			Port:     viper.GetInt("mysql.port"),
			User:     viper.GetString("mysql.user"),
			Password: viper.GetString("mysql.password"),
			DB:       viper.GetString("mysql.database"),
		}
	case "sqlite":
		Sqlite = sqliteConfig{
			DB: viper.GetString("sqlite.db"),
		}
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		MaxGoroutines = viper.GetInt("goroutine")
		BindMaxNum = viper.GetInt("bindmax")
		MaxErrTimes = viper.GetInt("errlimit")
		Notice = viper.GetString("notice")
		Admins = getAdmins()
	})
}
func getAdmins() []int64 {
	var result []int64
	admins := strings.Split(viper.GetString("admin"), ",")
	for _, v := range admins {
		id, _ := strconv.ParseInt(v, 10, 64)
		result = append(result, id)
	}
	return result
}
