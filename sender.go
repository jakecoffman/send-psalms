package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"

	"github.com/robfig/cron"
)

type EmailUser struct {
	Username    string
	Password    string
	EmailServer string
	Port        int
}

var password string
var bible map[string]map[string]map[string]string

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Enter next psalm to send and smtp password")
		return
	}
	password = os.Args[2]
	chapterStr := os.Args[1]
	chapterNum, err := strconv.Atoi(chapterStr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log.Println("Will send Psalm", chapterNum, "next")

	file, err := os.Open("bibles/ESV/ESV.json")
	if err != nil {
		log.Fatal(err)
	}
	json.NewDecoder(file).Decode(&bible)

	onCron := func() {
		subject, text := getEmailText(chapterStr)
		sendMail(subject, text)
		log.Println("Sent Psalms", chapterStr)
		chapterNum++
		if chapterNum > 150 {
			os.Exit(0)
		}
		chapterStr = strconv.Itoa(chapterNum)
	}

	c := cron.New()
	c.AddFunc("0 0 7 * * *", onCron)
	c.AddFunc("0 0 14 * * *", onCron)
	c.Start()

	select {} // sleep forever
}

func getEmailText(chapterStr string) (string, string) {
	psalms := bible["Psalms"]
	chapter := psalms[chapterStr]
	text := []string{}
	for i := 0; i < len(chapter); i++ {
		verse := strconv.Itoa(i + 1)
		text = append(text, chapter[verse])
	}

	subject := fmt.Sprintf("Psalms %v", chapterStr)
	return subject, strings.Join(text, " ")
}

func sendMail(subject string, content string) {
	emailUser := &EmailUser{
		"no-reply@coffshire.com",
		password,
		"smtp.gmail.com",
		587,
	}

	auth := smtp.PlainAuth("",
		emailUser.Username,
		emailUser.Password,
		emailUser.EmailServer,
	)

	err := smtp.SendMail(
		emailUser.EmailServer+":"+strconv.Itoa(emailUser.Port),
		auth,
		emailUser.Username,
		[]string{"jake@jakecoffman.com"},
		[]byte("From: no-reply@coffshire.com\r\n"+
			"To: jake@jakecoffman.com\r\n"+
			"Subject: "+subject+
			"\r\n\r\n"+
			content),
	)

	if err != nil {
		log.Fatal(err)
	}
}
