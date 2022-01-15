package mods

import (
	"bytes"
	"encoding/json"
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

type SendMessage struct {
	ChatId int    `json:"chat_id"`
	Text   string `json:"text"`
}

// Функция для отправки сообщений пользователю
func SendMsg(botUrl string, update Update, msg string) error {
	// Запись того, что и куда отправить
	botMessage := SendMessage{
		ChatId: update.Message.Chat.ChatId,
		Text:   msg,
	}

	// Запись сообщения в json
	buf, err := json.Marshal(botMessage)
	if err != nil {
		fmt.Println("Marshal json error: ", err)
		return err
	}

	// Отправка сообщения
	_, err = http.Post(botUrl+"/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		fmt.Println("SendMessage method error: ", err)
		return err
	}
	return nil
}

// Функция, которая ищет в ссылке значения fid tid и page_n
func parseCommand(str string) (int, string, string, int) {
	//str: 7777 https://www.banki.ru/forum/?PAGE_NAME=message&FID=77&TID=777777&PAGEN_1=7777#forum-message-list
	var fid, tid string
	var id, pagen int
	for i := 0; i < len(str); i++ {
		if str[i] == ' ' {
			id, _ = strconv.Atoi(str[:i])
			str = str[i+1:]
			break
		}
	}
	//str: https://www.banki.ru/forum/?PAGE_NAME=message&FID=77&TID=777777&PAGEN_1=7777#forum-message-list
	for i := 0; i < len(str); i++ {
		if str[i] == '&' {
			str = str[i+5:]
			break
		}
	}
	//str: 77&TID=777777&PAGEN_1=7777#forum-message-list
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

	//str: 77&TID=777777&PAGEN_1=7777
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

// Функция, проверяющая наличие новой страницы
func checkNewPage(botUrl string, update Update, fullUrl string) bool {
	// Запрос страницы
	resp, err := http.Get(fullUrl)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	// Если в ответ бот получил страницу с кодом 404, возвращает false
	if !strings.Contains(string(body), "<div class=\"ui-alert ui-alert--danger margin-top-default margin-bottom-default\">") {
		SendMsg(botUrl, update, "Новая страница!\n\n\n"+fullUrl)
		return true

	}
	return false
}

// Функция, проверяющая наличие новых сообщений
func checkNewMsg(botUrl string, update Update, fullUrl string, msgId int) int {
	// Запрос страницы
	resp, err := http.Get(fullUrl)
	var newMsgs int
	if err != nil {
		return -1
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1
	}
	// Поиск новых сообщений
	for newMsgs = 1; ; newMsgs++ {
		if strings.Contains(string(body), ">#"+strconv.Itoa(msgId+newMsgs)+"</a>") {
		} else {
			break
		}
	}
	return newMsgs - 1
}

// Функция, которая следит за тредом
func Check(botUrl string, update Update, str string) bool {
	// Получение значений из сообщений
	id, fid, tid, pagen := parseCommand(str)
	timeSinceLastCheck := time.Now().Unix()
	SendMsg(botUrl, update, "Начинаю следить за тредом")

	for {
		fmt.Println("\tChecking \t", id, fid, tid, pagen)
		newMsgs := checkNewMsg(botUrl, update, "https://www.banki.ru/forum/?PAGE_NAME=message&FID="+fid+"&TID="+tid+"&PAGEN_1="+strconv.Itoa(pagen), id)
		if newMsgs > 0 {
			SendMsg(botUrl, update, "Новых сообщений: "+strconv.Itoa(newMsgs-1)+"\n\n\n"+"https://www.banki.ru/forum/?PAGE_NAME=message&FID="+fid+"&TID="+tid+"&PAGEN_1="+strconv.Itoa(pagen))
			id += newMsgs
			timeSinceLastCheck = time.Now().Unix()
		} else {
			if checkNewPage(botUrl, update, "https://www.banki.ru/forum/?PAGE_NAME=message&FID="+fid+"&TID="+tid+"&PAGEN_1="+strconv.Itoa(pagen+1)) {
				pagen++
				id++
				timeSinceLastCheck = time.Now().Unix()
			}
		}

		fmt.Println("Sleep \t", id, fid, tid, pagen)
		time.Sleep(time.Minute * 10)

		if time.Now().Unix()-timeSinceLastCheck > 86400 {
			SendMsg(botUrl, update, "24 Часа прошло с последнего сообщения в треде:\n\n"+"https://www.banki.ru/forum/?PAGE_NAME=message&FID="+fid+"&TID="+tid+"&PAGEN_1="+strconv.Itoa(pagen)+"\n\n перестаю за ним следить")
			return true
		}

	}
}

// Функция инициализации конфига (токенов)
func InitConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}
