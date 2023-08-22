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
		"Please read the documentation during your binding process: [Click to view documentation](https://note.toshiki.dev/application/toshiki-e5subot)",
		tb.ModeMarkdown,
	)

	bot.Send(m.Chat,
		"⚠ Please reply in the following format `client_id(space)client_secret`",
		&tb.SendOptions{ParseMode: tb.ModeMarkdown,
			ReplyMarkup: &tb.ReplyMarkup{ForceReply: true}},
	)

	UserStatus[m.Chat.ID] = StatusBind1
	UserClientId[m.Chat.ID] = m.Text
}

func bBind1(m *tb.Message) {
	if !m.IsReply() {
		bot.Send(m.Chat, "Please bind through replying to the messages") // Please bind through replying to the messages
		return
	}
	tmp := strings.Split(m.Text, " ")
	if len(tmp) != 2 {
		bot.Send(m.Chat, "⚠ Wrong format inputted")
		return
	}
	id := tmp[0]
	secret := tmp[1]
	bot.Send(m.Chat,
		fmt.Sprintf("👉 Please authorize your account - [click to login for granting access](%s)", microsoft.GetAuthURL(id)),
		tb.ModeMarkdown,
	)

	bot.Send(m.Chat,
		"⚠ Please reply the full fallback back url from your address bar with format of `http://localhost/......(space)alias` for convenient management purposes",
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
		bot.Send(m.Chat, "⚠ Wrong format inputted")
		return
	}
	if len(srv_client.GetClients(m.Chat.ID)) == config.BindMaxNum {
		bot.Send(m.Chat, "⚠ You have reached the maximum accoutn binding limits, please consider remove exesscive or any unused accounts to contiue a new bind")
		return
	}
	bot.Send(m.Chat, "⚠ Account binding in process, please standy by for a bot response...")

	tmp := strings.Split(m.Text, " ")
	if len(tmp) != 2 {
		bot.Send(m.Chat, "⚠ Wrong format inputted")
	}
	code := util.GetURLValue(tmp[0], "code")
	alias := tmp[1]

	id := UserClientId[m.Chat.ID]
	secret := UserClientSecret[m.Chat.ID]

	refresh, err := microsoft.GetTokenWithCode(id, secret, code)
	if err != nil {
		bot.Send(m.Chat, fmt.Sprintf("Failed to fetch a ResponseToken, please restart the binding process ERROR:%s", err))
		return
	}
	bot.Send(m.Chat, "🎉 Successfully obtained RefreshToken, congratulations")

	refresh, info, err := microsoft.GetUserInfo(id, secret, refresh)
	if err != nil {
		bot.Send(m.Chat, fmt.Sprintf("Failed to fetch user information ERROR:%s", err))
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
		bot.Send(m.Chat, "⚠ This certain application or account is already successfully bound to the bot, failed to rebind")
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
		bot.Send(m.Chat, "⚠ Failed write user data into database")
		return
	}

	bot.Send(m.Chat, "✨ Congratulations, account bound successfully; happy using!")
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
		fmt.Sprintf("⚠ Please select an account ot unbind\n\nCurrent bound accounts: %d/%d", len(srv_client.GetClients(m.Chat.ID)), config.BindMaxNum),
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
		bot.Send(c.Message.Chat, "⚠ Failed to unbind, please recheck your configuration")
		return
	}
	bot.Send(c.Message.Chat, "✨ Successfully unbind, you are welcomed to reuse the bot at anytime in future again")
	bot.Respond(c)
}
