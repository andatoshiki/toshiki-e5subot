package config

import (
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	LogBasePath    string = "./log/"
	WelcomeContent string = "Welcome to toshiki-e5subot!"
	HelpContent    string = `
	🤖 Toshiki's E5Subot is a Microsoft developer E5 account renwal automation program integrated with Telegram bot APIs via calling Microsoft Graph APIs.

	Below are the commonly used commands:

	/start - Send welcome messages
	/my - View bound accouts information details
	/bind - Bind new accounts
	/unbind - Ubind existing accounts
	/export - Export account information details (JSON)
	/help - Bot help guides
	/task - Manual trigger bot for renwal API calling (admins only)
	/log - Fetch bot historical logs for debug (admins only)
	
	Github: github.com/andatoshiki/toshiki-e5subot
	Docs: note.toshiki.dev/application/toshiki-e5subot
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
