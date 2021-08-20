package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
	"github.com/PaulSonOfLars/gotgbot/handlers/Filters"
	"github.com/wabarc/go-anonfile"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/sqltocsv"
)

var log = zap.NewProductionEncoderConfig()

var logger = zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(log), os.Stdout, zap.InfoLevel))
var db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/greencoin")

func main() {
	// fmt.println("Hi")
	// fmt.println(u.EffectiveMessage)
	updater, err := gotgbot.NewUpdater(logger, "1900533231:AAE1Yb6HAkiziDGiKZIL0RObwqhDVxG5JJ4")

	if err != nil {
		logger.Panic("UPDATER FAILED TO START")
	}
	logger.Sugar().Info("UPDATER STARTED SUCCESFULLY")
	updater.StartCleanPolling()
	updater.Dispatcher.AddHandler(handlers.NewCommand("start", start))
	updater.Dispatcher.AddHandler(handlers.NewCommand("addtwitter", addtwit))
	updater.Dispatcher.AddHandler(handlers.NewCommand("getcsv", getcsv))

	updater.Dispatcher.AddHandler(handlers.NewCommand("updatewallet", updateAddress))

	updater.Dispatcher.AddHandler(handlers.NewMessage(Filters.Text, verify))

	updater.Idle()

}
func getcsv(b ext.Bot, u *gotgbot.Update) error {
	rows, _ := db.Query("SELECT id,userid,twitter_handle,tg_usr_name,referals,address,balance FROM users WHERE verified=1")
	var chat_id = u.EffectiveChat.Id
	err := sqltocsv.WriteFile("go.csv", rows)
	if err != nil {
		panic(err)
	}
	var path = "go.csv"
	if urgl, err := anonfile.NewAnonfile(nil).Upload(path); err != nil {
		logger.Panic("Error While Uploading")
	} else {
		fmt.Print(urgl.Full)
		// var url = fmt.Sprintf("%s", urgl)
		_ = chat_id
		b.SendMessage(chat_id, urgl.Full())
	}
	return nil
}
func updateAddress(b ext.Bot, u *gotgbot.Update) error {
	textg := u.EffectiveMessage.Text
	var text = strings.Fields(textg)
	var chat_id = u.EffectiveChat.Id
	var res, err = b.GetChatMember(os.Getenv("group_id"), chat_id)

	if res.Status != "left" {
		var res, err = b.GetChatMember(os.Getenv("channel_id"), chat_id)
		_ = err

		if res.Status != "left" {
			type Users struct {
				twitter_handle string `json:"twitter_handle"`
				referered_by   int    `json:"referered_by"`
				address        string `json:"address"`
			}

			var user Users
			err := db.QueryRow("SELECT twitter_handle,referered_by,address FROM users WHERE userid = ?", chat_id).Scan(&user.twitter_handle, &user.referered_by, &user.address)
			_ = err
			if user.twitter_handle != "none" {
				if len(text) > 1 {
					if user.address == "none" {
						if text[1] != "none" {
							db.Query("UPDATE users set balance=balance+100000,address='"+text[1]+"' WHERE userid= ?", chat_id)
							b.SendMessage(chat_id, "Address Updated Successfully!!\n\n Balance increased by 100000 GREEN.")
						}
					}

					if user.address != "none" {
						db.Query("UPDATE users set address='"+text[1]+"' WHERE userid= ?", chat_id)
						b.SendMessage(chat_id, "Address Updated Successfully!!")
					}

				}
			}
			if user.twitter_handle == "none" {
				b.SendMessage(chat_id, "You need to add your twitter first")
			}
		}
	}
	_ = err
	if res.Status == "left" {
		var res, err = b.GetChatMember(-1001153791846, chat_id)
		_ = err

		if res.Status == "left" {
			b.SendMessage(chat_id, "✖ Please join both of our telegram group and channel.")
		}
		b.SendMessage(chat_id, "✖ Please join both of our telegram group and channel.")

	}
	return nil
}
func addtwit(b ext.Bot, u *gotgbot.Update) error {
	textg := u.EffectiveMessage.Text
	var text = strings.Fields(textg)
	var chat_id = u.EffectiveChat.Id
	var res, err = b.GetChatMember(-1001153791846, chat_id)

	if res.Status != "left" {
		var res, err = b.GetChatMember(-1001153791846, chat_id)
		_ = err

		if res.Status != "left" {
			type Users struct {
				twitter_handle string `json:"twitter_handle"`
				referered_by   int    `json:"referered_by"`
			}

			var user Users
			err := db.QueryRow("SELECT twitter_handle,referered_by FROM users WHERE userid = ?", chat_id).Scan(&user.twitter_handle, &user.referered_by)
			_ = err
			if user.twitter_handle == "none" {
				if len(text) > 1 {
					var twitter = strings.Replace(text[1], "@", "", -1)
					db.Query("UPDATE users SET twitter_handle='https://twitter.com/" + twitter + "' WHERE userid=" + strconv.Itoa(chat_id) + "")
					b.SendMessage(chat_id, "Twitter Username Updated successfully. Now pls set your wallet address(Bep20) by /updatewallet 0x4d......")
					db.Query("UPDATE users SET balance=balance+25000  WHERE userid=" + strconv.Itoa(user.referered_by) + "")
					b.SendMessage(user.referered_by, "🥳 Balance increased by 25000 GREEN.")

				}
			}
		}
	}
	_ = err
	if res.Status == "left" {
		var res, err = b.GetChatMember(-1001153791846, chat_id)
		_ = err

		if res.Status == "left" {
			b.SendMessage(chat_id, "✖ Please join both of our telegram group and channel.")
		}
		b.SendMessage(chat_id, "✖ Please join both of our telegram group and channel.")

	}
	return nil
}
func info(b ext.Bot, u *gotgbot.Update) error {

	logger.Sugar().Info(u.EffectiveMessage.Text)
	return nil
}
func verify(b ext.Bot, u *gotgbot.Update) error {
	var chat_id int = u.EffectiveChat.Id
	if u.EffectiveMessage.Text == "✅ Submit Info" {
		var res, err = b.GetChatMember(-1001153791846, chat_id)
		if res.Status != "left" {
			var res, err = b.GetChatMember(-1001153791846, chat_id)
			_ = err

			if res.Status != "left" {
				type Users struct {
					twitter_handle string `json:"twitter_handle"`
				}
				var user Users
				err := db.QueryRow("SELECT twitter_handle FROM users WHERE userid = ?", chat_id).Scan(&user.twitter_handle)
				_ = err
				if user.twitter_handle == "none" {
					var g, err = b.SendMessage(chat_id, "Follow us on Twitter : https://twitter.com/Greenlifecoin\n\n Submit your username(without @) after following us by using /addtwitter command.\n\n For example, /addtwitter username")
					_ = g
					_ = err
				}
			}
		}
		_ = err
		if res.Status == "left" {
			var res, err = b.GetChatMember(-1001153791846, chat_id)
			_ = err

			if res.Status == "left" {
				b.SendMessage(chat_id, "✖ Please join both of our telegram group and channel.")
			}
			b.SendMessage(chat_id, "✖ Please join both of our telegram group and channel.")

		}
		return err
	}
	if u.EffectiveMessage.Text == "💻 Airdrop Info" {
		b.SendMessage(u.EffectiveChat.Id, "Airdrop Rewards 100k Green (10$)\n\nReferrals Rewards 25k Green (2.5$)\n\nTop 10 Referrars will get 100 Busd Each.\n\nAirdrop Will be end 20th November\n\nDistribution Date: 1st December 2021 \n\nnSocial Media Links:\n\nChannel: https://t.me/GreenLifeCoinANN\n\nnGroup: https://t.me/Green_Life_Coin\n\nTwitter: https://mobile.twitter.com/Greenlifecoin \n\nFacebook : https://m.facebook.com/GreenLifeCoin.Organization/\n\nDiscord: https://discord.gg/2GWjpgp\n\nWebsite: https://greenlifecoin.com\n\n")
		return nil
	}
	if u.EffectiveMessage.Text == "🙌 Referals" {
		type User struct {
			balance  int    `json:"balance"`
			referals int    `json:"referals"`
			address  string `json:"address"`
		}
		var user User
		err := db.QueryRow("SELECT balance,referals,address FROM users WHERE userid = ?", chat_id).Scan(&user.balance, &user.referals, &user.address)
		_ = err

		b.SendMessageHTML(u.EffectiveChat.Id, "You have "+strconv.Itoa(user.referals)+" referrals, and have total balance of "+strconv.Itoa(user.balance)+" <b>GREEN</b>.\n\nWallet Address: <pre>"+user.address+"</pre>\n\nTo refer people, send them to:\n\n<pre>https://t.me/GreenLifeCoin_AirdropBot?start="+strconv.Itoa(chat_id)+"</pre>\n\nYou will earn 25k Green after your friends joins the airdrop and submits his info.")
		return nil
	}
	var is_user bool
	is_user = check_user(chat_id)
	if is_user != true {
		return nil
	}
	type User struct {
		id       int `json:"id"`
		verified int `json:"verified"`
		get_calc int `json:"get_calc"`
	}
	var user User
	err := db.QueryRow("SELECT id,verified,get_calc FROM users WHERE userid = ?", chat_id).Scan(&user.id, &user.verified, &user.get_calc)
	_ = err
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	if user.verified == 0 {
		var sum int
		var msg string = u.EffectiveMessage.Text
		sum, err = strconv.Atoi(msg)
		if sum == user.get_calc {
			db.Query("Update users SET verified=1 where userid= ?", chat_id)
			type sendMessageReqBody struct {
				ChatID   int      `json:"chat_id"`
				Text     string   `json:"text"`
				KeyBoard []string `json:reply_markup`
			}
			reqBody := &sendMessageReqBody{
				ChatID: chat_id,
				Text:   "☑ Human Verification Successful!!",
			}
			reqBytes, err := json.Marshal(reqBody)
			res, err := http.Post("https://api.telegram.org/bot1900533231:AAE1Yb6HAkiziDGiKZIL0RObwqhDVxG5JJ4/sendMessage?reply_markup={%22keyboard%22:[[{%22text%22:%22💻 Airdrop Info%22},{%22text%22:%22🙌 Referals%22}],[{%22text%22:%22✅ Submit Info%22}]],%22resize_keyboard%22:true}", "application/json", bytes.NewBuffer(reqBytes))
			_ = res
			_ = err
		}
		if sum != user.get_calc {
			b.SendMessage(chat_id, "✖ Human Verification Failed Try again!!")
		}
		return nil

	}
	return nil
}
func check_user(chat_id int) bool {
	var status bool
	status = true

	if err != nil {
		logger.Panic(err.Error())
	}
	results, err := db.Query("SELECT * FROM users WHERE userid=" + strconv.Itoa(chat_id) + "")
	if err != nil && err != sql.ErrNoRows {
		status = false // proper error handling instead of panic in your app
	}
	if results.Next() {
		//exists
		status = true

	} else {
		status = false // proper error handling instead of panic in your app

	}
	_ = results
	return status
}
func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}
func start(b ext.Bot, u *gotgbot.Update) error {
	textg := u.EffectiveMessage.Text
	var text = strings.Fields(textg)
	_ = text
	fmt.Println(text)
	var is_user bool
	b.SendMessage(u.EffectiveChat.Id, "Greenlife Coin™ ( GREEN ) is the world’s leading decentralized digital currency funding system. We provide the most secure and convenient way to fund, create, buy, sell, explore and trade various exchanges – to anyone, anywhere in the world.\n\nAs we continue to grow and deliberate more on this project, we decided to become a fully decentralized organization. GREEN  aims to empower her community and encourage active involvement by giving them the power to fulfill their project requirements on environmental conservation.\n\nMeaning, GREEN tokens can be earned through active engagements on the platform. ")
	var user_id = u.EffectiveChat.Id
	var user_name string = u.EffectiveChat.Username
	_ = user_name
	is_user = check_user(user_id)
	_ = is_user
	var a, c, sum int = 0, 0, 0
	_ = a + c + sum
	logger.Sugar().Info(is_user)
	if is_user == false {
		a = 0
		c = 0
		sum = 0
		a = rangeIn(100, 999)
		c = rangeIn(100, 999)
		sum = a + c
		var sms string
		sms = "➡️Before we start the airdrop, please prove you are human by answering the question below.\n\nPlease answer: " + strconv.Itoa(a) + " + " + strconv.Itoa(c) + " ="
		b.SendMessage(user_id, sms)
		// b.SendMessage(u.EffectiveChat.Id, u.EffectiveChat.Id)
		if len(text) > 1 {
			var referer int
			if _, err := strconv.Atoi(text[1]); err == nil {
				referer, err = strconv.Atoi(text[1])
				if check_user(referer) != false && referer != user_id {
					db.Query("Update users SET referals=referals+1 where userid= ?", referer)
					b.SendMessage(referer, "🥳 New users joined the bot using your referral link!! Bonus will be added after he/she completes the steps and submits the airdrop.")
					results, err := db.Query("INSERT INTO users (userid,balance,verified,tg_usr_name,joined_at,referals,joined,get_calc,referered_by) values(" + strconv.Itoa(user_id) + ",0,0,'" + user_name + "'," + strconv.FormatInt(time.Now().Unix(), 10) + ",0,0," + strconv.Itoa(sum) + "," + strconv.Itoa(referer) + ") ")
					_ = results
					if err != nil {
						logger.Panic(err.Error())
					}
					return nil
				}
			}
			if check_user(referer) != true {
				results, err := db.Query("INSERT INTO users (userid,balance,verified,tg_usr_name,joined_at,referals,joined,get_calc) values(" + strconv.Itoa(user_id) + ",0,0,'" + user_name + "'," + strconv.FormatInt(time.Now().Unix(), 10) + ",0,0," + strconv.Itoa(sum) + ") ")
				_ = results
				if err != nil {
					logger.Panic(err.Error())
				}
			}
		}
		is_user = check_user(user_id)
		if is_user == false {

			results, err := db.Query("INSERT INTO users (userid,balance,verified,tg_usr_name,joined_at,referals,joined,get_calc) values(" + strconv.Itoa(user_id) + ",0,0,'" + user_name + "'," + strconv.FormatInt(time.Now().Unix(), 10) + ",0,0," + strconv.Itoa(sum) + ") ")
			_ = results
			if err != nil {
				logger.Panic(err.Error())
			}
		}
	}

	return nil
}
