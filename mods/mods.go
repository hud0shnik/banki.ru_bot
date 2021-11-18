package mods

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type TelegramResponse struct {
	Result []Update `json:"result"`
}

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ChatId int `json:"id"`
}

func InitConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}

func parseCommand(str string) (int, string, string, int) {
	//7777 https://www.banki.ru/forum/?PAGE_NAME=message&FID=77&TID=777777&PAGEN_1=7777#forum-message-list
	var fid, tid string
	var id, pagen int
	for i := 0; i < len(str); i++ {
		if str[i] == ' ' {
			id, _ = strconv.Atoi(str[:i])
			str = str[i+1:]
			break
		}
	}
	//https://www.banki.ru/forum/?PAGE_NAME=message&FID=77&TID=777777&PAGEN_1=7777#forum-message-list
	for i := 0; i < len(str); i++ {
		if str[i] == '&' {
			str = str[i+5:]
			break
		}
	}
	//77&TID=777777&PAGEN_1=7777#forum-message-list
	for i := 0; i < len(str); i++ {
		if str[i] == '&' {
			break
		}
		fid += string(str[i])
	}
	if strings.Contains(str, "#") {
		for i := 0; i < len(str); i++ {
			if str[i] == '#' {
				str = str[:i]
				break
			}
		}
	}

	str = str[len(fid)+5:]
	for i := 0; i < len(str); i++ {
		if str[i] == '&' {
			break
		}
		tid += string(str[i])
	}

	pagen, _ = strconv.Atoi(str[len(fid)+13:])
	return id, fid, tid, pagen
}

func checkNewMsg(botUrl string, update Update, fullUrl string, msgId int) bool {
	resp, err := http.Get(fullUrl)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	if strings.Contains(string(body), ">#"+strconv.Itoa(msgId+1)+"</a>") {
		SendMsg(botUrl, update, "Новое сообщение!\n\n\n"+fullUrl)

		return true
	}
	return false
}

func checkNewPage(botUrl string, update Update, fullUrl string) bool {
	resp, err := http.Get(fullUrl)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	if !strings.Contains(string(body), "<div class=\"ui-alert ui-alert--danger margin-top-default margin-bottom-default\">") {
		SendMsg(botUrl, update, "Новая страница!\n\n\n"+fullUrl)
		return true

	}
	return false

}

func Check(botUrl string, update Update, str string) bool {

	id, fid, tid, pagen := parseCommand(str)
	timeSinceLastCheck := time.Now().Unix()
	SendMsg(botUrl, update, "Начинаю следить за тредом")

	for {
		fmt.Println("\tChecking new msg....\n", id, fid, tid, pagen)
		if checkNewMsg(botUrl, update, "https://www.banki.ru/forum/?PAGE_NAME=message&FID="+fid+"&TID="+tid+"&PAGEN_1="+strconv.Itoa(pagen), id) {
			id++
			timeSinceLastCheck = time.Now().Unix()
		} else {
			if checkNewPage(botUrl, update, "https://www.banki.ru/forum/?PAGE_NAME=message&FID="+fid+"&TID="+tid+"&PAGEN_1="+strconv.Itoa(pagen+1)) {
				pagen++
				id++
				timeSinceLastCheck = time.Now().Unix()
			}
		}

		fmt.Println("sleep...")
		time.Sleep(time.Minute * 10)

		if time.Now().Unix()-timeSinceLastCheck > 86400 {
			SendMsg(botUrl, update, "24 часа прошло с последнего сообщения в треде:\n\n"+"https://www.banki.ru/forum/?PAGE_NAME=message&FID="+fid+"&TID="+tid+"&PAGEN_1="+strconv.Itoa(pagen)+"\n\n перестаю за ним следить")
			return true
		}

	}

}
