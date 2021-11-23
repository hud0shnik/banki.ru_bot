package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"tgBot/mods"

	"github.com/spf13/viper"
)

func main() {
	err := mods.InitConfig()
	if err != nil {
		log.Println("Config error: ", err)
		return
	}
	botUrl := "https://api.telegram.org/bot" + viper.GetString("token")
	offSet := 0
	for {
		updates, err := getUpdates(botUrl, offSet)
		if err != nil {
			log.Println("Something went wrong: ", err)
		}
		for _, update := range updates {
			respond(botUrl, update)
			offSet = update.UpdateId + 1
		}
		fmt.Println(updates)
	}
}

func getUpdates(botUrl string, offset int) ([]mods.Update, error) {
	resp, err := http.Get(botUrl + "/getUpdates?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var restResponse mods.TelegramResponse
	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		return nil, err
	}
	return restResponse.Result, nil
}

//	https://core.telegram.org/bots/api#using-a-local-bot-api-server
func respond(botUrl string, update mods.Update) error {

	msg := update.Message.Text

	if msg == "/help" || msg == "/start" {
		mods.SendMsg(botUrl, update, "Чтобы подписать бота на тред, отправьте айди вашего сообщения и ссылку на сам тред в формате\n"+
			"\"777 https://www.banki.ru/forum/?PAGE_NAME=read&FID=77&TID=77777&PAGEN_1=777#forum-message-list\"\n"+
			"когда появятся новые сообщения, бот вас оповестит\n")
		return nil
	}

	if strings.Contains(msg, "https://www.banki.ru/forum/?PAGE_NAME=") {
		go mods.Check(botUrl, update, msg)
		return nil
	}

	mods.SendMsg(botUrl, update, "I don't understand, sorry")
	return nil
}
