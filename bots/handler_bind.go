package bots

import (
	"fmt"
	"github.com/andatoshiki/toshiki-e5subot/config"
	"github.com/andatoshiki/toshiki-e5subot/model"
	"github.com/andatoshiki/toshiki-e5subot/pkg/microsoft"
	"github.com/andatoshiki/toshiki-e5subot/service/srv_client"
	"github.com/andatoshiki/toshiki-e5subot/util"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
	"strconv"
	"strings"
)

func bBind(m *tb.Message) {
	bot.Send(m.Chat,
		"👉 客官注册前请先查看教程哦: [查看教程](https://telegra.ph/%E4%BF%8A%E6%A8%B9%E3%81%AEE5subot%E4%BD%BF%E7%94%A8%E6%95%99%E7%A8%8B-08-02)",
		tb.ModeMarkdown,
	)

	bot.Send(m.Chat,
		"⚠ 请使用如下格式回复哦 `client_id(空格)client_secret`",
		&tb.SendOptions{ParseMode: tb.ModeMarkdown,
			ReplyMarkup: &tb.ReplyMarkup{ForceReply: true}},
	)

	UserStatus[m.Chat.ID] = StatusBind1
	UserClientId[m.Chat.ID] = m.Text
}

func bBind1(m *tb.Message) {
	if !m.IsReply() {
		bot.Send(m.Chat, "⚠ 笨蛋! 请通过回复方式绑定! x_x")
		return
	}
	tmp := strings.Split(m.Text, " ")
	if len(tmp) != 2 {
		bot.Send(m.Chat, "⚠ 笨蛋! 格式错啦! >_<")
		return
	}
	id := tmp[0]
	secret := tmp[1]
	bot.Send(m.Chat,
		fmt.Sprintf("👉 请授权账户哦： [点击直达](%s)", microsoft.GetAuthURL(id)),
		tb.ModeMarkdown,
	)

	bot.Send(m.Chat,
		"⚠ 请回复`http://localhost/......(空格)别名`的格式哦~ (用于管理)",
		&tb.SendOptions{ParseMode: tb.ModeMarkdown,
			ReplyMarkup: &tb.ReplyMarkup{ForceReply: true},
		},
	)
	UserStatus[m.Chat.ID] = StatusBind2
	UserClientId[m.Chat.ID] = id
	UserClientSecret[m.Chat.ID] = secret
}

func bBind2(m *tb.Message) {
	if !m.IsReply() {
		bot.Send(m.Chat, "⚠ 笨蛋! 格式错啦! >_<")
		return
	}
	if len(srv_client.GetClients(m.Chat.ID)) == config.BindMaxNum {
		bot.Send(m.Chat, "⚠ 已经达到最大可绑定数啦! 在这样下去 我...我要坏掉了呜呜 ≧.≦")
		return
	}
	bot.Send(m.Chat, "正在绑定中哦 请耐心等待.......")

	tmp := strings.Split(m.Text, " ")
	if len(tmp) != 2 {
		bot.Send(m.Chat, "😥 笨蛋! 格式错啦! >_<")
	}
	code := util.GetURLValue(tmp[0], "code")
	alias := tmp[1]

	id := UserClientId[m.Chat.ID]
	secret := UserClientSecret[m.Chat.ID]

	refresh, err := microsoft.GetTokenWithCode(id, secret, code)
	if err != nil {
		bot.Send(m.Chat, fmt.Sprintf("呜呜 无法获取RefreshToken ERROR:%s", err))
		return
	}
	bot.Send(m.Chat, "🎉 Token获取成功哦! ^_^")

	refresh, info, err := microsoft.GetUserInfo(id, secret, refresh)
	if err != nil {
		bot.Send(m.Chat, fmt.Sprintf("无法获取用户信息呜呜 坏掉啦 ERROR:%s", err))
		return
	}
	c := &model.Client{
		TgId:         m.Chat.ID,
		RefreshToken: refresh,
		MsId:         util.Get16MD5Encode(gjson.Get(info, "id").String()),
		Alias:        alias,
		ClientId:     id,
		ClientSecret: secret,
		Other:        "",
	}

	if srv_client.IsExist(c.TgId, c.ClientId) {
		bot.Send(m.Chat, "⚠ 笨蛋! 该应用已经绑定过了 无需重复绑定 我很聪明的!")
		return
	}

	bot.Send(m.Chat,
		fmt.Sprintf("ms_id：%s\nuserPrincipalName：%s\ndisplayName：%s",
			c.MsId,
			gjson.Get(info, "userPrincipalName").String(),
			gjson.Get(info, "displayName").String(),
		),
	)

	if err = srv_client.Add(c); err != nil {
		bot.Send(m.Chat, "😥 用户写入数据库失败啦")
		return
	}

	bot.Send(m.Chat, "✨ 恭喜恭喜! 绑定成功啦! 祝您使用愉快!")
	delete(UserStatus, m.Chat.ID)
	delete(UserClientId, m.Chat.ID)
	delete(UserClientSecret, m.Chat.ID)
}

func bUnBind(m *tb.Message) {
	var inlineKeys [][]tb.InlineButton
	clients := srv_client.GetClients(m.Chat.ID)

	for _, u := range clients {
		inlineBtn := tb.InlineButton{
			Unique: "unbind" + strconv.Itoa(u.ID),
			Text:   u.Alias,
			Data:   strconv.Itoa(u.ID),
		}
		bot.Handle(&inlineBtn, bUnBindInlineBtn)
		inlineKeys = append(inlineKeys, []tb.InlineButton{inlineBtn})
	}

	bot.Send(m.Chat,
		fmt.Sprintf("⚠ 请选择一个账户将其解绑\n\n当前绑定数: %d/%d", len(srv_client.GetClients(m.Chat.ID)), config.BindMaxNum),
		&tb.ReplyMarkup{InlineKeyboard: inlineKeys},
	)
}
func bUnBindInlineBtn(c *tb.Callback) {
	id, _ := strconv.Atoi(c.Data)
	if err := srv_client.Del(id); err != nil {
		zap.S().Errorw("failed to delete db data",
			"error", err,
			"id", c.Data,
		)
		bot.Send(c.Message.Chat, "⚠ 解绑失败! x_x 看看是不是哪里出问题啦?")
		return
	}
	bot.Send(c.Message.Chat, "✨ 解绑成功! 欢迎下次光临哦~")
	bot.Respond(c)
}
