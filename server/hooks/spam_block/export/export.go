package main

import (
	"fmt"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/hooks/spam_block/tools"
	"github.com/Jinnrry/pmail/models"
	"github.com/spf13/cast"
	"os"
)

func getType(emailId int) int {
	var ue models.UserEmail
	_, err := db.Instance.Table(&ue).Where("email_id = ?", emailId).Limit(1).Get(&ue)
	if err != nil {
		fmt.Println(err)
	}
	if ue.Status == 3 {
		return 2
	}

	if ue.Status == 5 {
		return 1
	}

	return 0
}

func main() {
	args := os.Args

	var id int
	if len(args) >= 2 {
		id = cast.ToInt(args[1])
	}

	config.Init()
	err := db.Init("test")
	if err != nil {
		panic(err)
	}
	fmt.Println(config.Instance.DbDSN)

	fmt.Println("文件第一列是分类，0表示正常邮件，1表示垃圾邮件，2表示诈骗邮件")

	var start int
	file, err := os.OpenFile("data.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	for {
		var emails []models.Email
		if id > 0 {
			db.Instance.Table(&models.Email{}).Where("id = ?", id).OrderBy("id").Find(&emails)
		} else {
			db.Instance.Table(&models.Email{}).Where("id > ?", start).OrderBy("id").Find(&emails)
		}
		if len(emails) == 0 {
			break
		}
		for _, email := range emails {
			start = email.Id
			content := tools.Trim(tools.TrimHtml(email.Html.String))
			if content == "" {
				content = tools.Trim(email.Text.String)
			}
			_, err = file.WriteString(fmt.Sprintf("%d \t%s %s\n", getType(email.Id), email.Subject, content))
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Printf("0 \t%s %s\n", email.Subject, trim(trimHtml(email.Html.String)))
		}
		if id > 0 {
			break
		}

	}

}
