package main

import (
	"code.google.com/p/gopass"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/scorredoira/email"
	"io/ioutil"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

var (
	Conf   tConf
	Attach []string

	O = string(os.PathSeparator)

	config = "mailer.json"
)

type tConf struct {
	Server  string   `json:"server"`
	Port    string   `json:"port"`
	From    string   `json:"from"`
	To      string   `json:"to"`
	User    string   `json:"user"`
	Pass    string   `json:"pass"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
	Attach  []string `json:"attach"`
}

func init() {
	files := flag.String("f", "", "one or list of files")
	dir := flag.String("d", "", "dir with files")
	text := flag.String("t", "", "text")
	subject := flag.String("s", "", "subject")
	flag.Parse()

	if _, err := os.Stat(config); os.IsNotExist(err) {
		if _, err := os.Stat("~/.mailer.json"); os.IsNotExist(err) {
			config = "~/.mailer.json"
		}
	}
	conf, err := ioutil.ReadFile(config)
	if err != nil {
		fmt.Println("eror read " + config)
		fmt.Println(err.Error())
		usage()
	}
	err = json.Unmarshal(conf, &Conf)
	if err != nil {
		fmt.Println("eror read " + config)
		fmt.Println(err.Error())
		usage()
	}

	if len(flag.Args()) >= 1 {
		args := flag.Args()
		Conf.To = args[0]
	}

	//////

	if len(Conf.To) == 0 {
		usage()
	}

	if len(*files) > 0 {
		Conf.Attach = strings.Split(*files, ",")
	}

	if len(*dir) > 0 {
		Conf.Attach = []string{*dir}
	}

	if len(*text) > 0 {
		Conf.Text = *text
	}

	if len(*subject) > 0 {
		Conf.Subject = *subject
	}

	if len(Conf.Pass) == 0 {
		Conf.Pass, err = gopass.GetPass("pass:")
		checkerr(err)
	}

}

func usage() {
	fmt.Println("usage: " + os.Args[0] + " [option] email")
	fmt.Println("	-d	dir with files")
	fmt.Println("	-f	one or list of files sep \",\"")
	fmt.Println("	-t	text")
	fmt.Println("	-s	subject")
	os.Exit(0)
}

func getfiles(filelist []string) []string {
	var files []string
	if len(files) > 10 {
		fmt.Println("too many files")
		fmt.Println(files)
		os.Exit(0)
	}
	for _, path := range filelist {
		stat, err := os.Stat(path)
		if os.IsNotExist(err) {
			fmt.Println("file not found: " + path)
			continue
		}
		if stat.IsDir() {
			filesInDir, err := ioutil.ReadDir(path)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			var f []string
			for i := 0; i < len(filesInDir); i++ {
				f = append(f, path+O+filesInDir[i].Name())
			}

			dirfiles := getfiles(f)
			files = append(files, dirfiles...)
		} else {
			files = append(files, path)
		}
	}
	return files
}

func main() {
	mail := email.NewMessage(Conf.Subject, Conf.Text)
	mail.From = Conf.From
	mail.To = []string{Conf.To}

	if len(Conf.Attach) > 0 {
		files := getfiles(Conf.Attach)

		for i := 0; i < len(files); i++ {
			mail.Attach(files[i])
		}
	}

	err := email.Send(Conf.Server+":"+Conf.Port, smtp.PlainAuth("", Conf.User, Conf.Pass, Conf.Server), mail)
	checkerr(err, "message not send", "message send to:"+Conf.To+" files:"+strconv.Itoa(len(Conf.Attach)))
}

func isset(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func checkerr(err error, msg ...string) {
	if err != nil && len(msg) >= 1 && len(msg[0]) > 0 {
		fmt.Println(msg[0])
	}
	if err == nil && len(msg) >= 2 && len(msg[1]) > 0 {
		fmt.Println(msg[1])
	}

	if err != nil {
		panic(err)
	}
}
