package main

import (
	"bufio"
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

const URL = `<html><body><a href="http://www.esvbible.org/%s+%s">ESV</a>`
const PSALMS = "Psalms"

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: %q [next chapter to send] [smtp password]")
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

	// populate initial list of "tos"
	tos, err := readLines("send-to.txt")
	if err != nil {
		log.Fatal(err)
	}

	onCron := func() {
		text := fmt.Sprintf(URL, PSALMS, chapterStr)
		subject := fmt.Sprintf("%s %s", PSALMS, chapterStr)

		// see if there are updated tos
		newTos, err := readLines("send-to.txt")
		if err != nil {
			log.Println("Warning: can't find send-to.txt")
		} else {
			tos = newTos
		}
		sendMail(subject, text, tos)
		log.Println("Sent Psalms", chapterStr)
		chapterNum++
		if chapterNum > 150 {
			os.Exit(0)
		}
		chapterStr = strconv.Itoa(chapterNum)
	}

	c := cron.New()
	c.AddFunc("0 0 7 * * *", onCron)
	c.AddFunc("0 0 19 * * *", onCron)
	c.Start()

	select {} // sleep forever
}

func sendMail(subject, content string, tos []string) {
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
			"To: "+strings.Join(tos, ",")+"\r\n"+
			"Subject: "+subject+"\r\n"+
			"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"+
			"\r\n\r\n"+
			content),
	)

	if err != nil {
		log.Fatal(err)
	}
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
