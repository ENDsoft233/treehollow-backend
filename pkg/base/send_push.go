package base

import (
	"bytes"
	"encoding/json"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"log"
	"net/http"
	"treehollow-v3-backend/pkg/model"
)

func PreProcessPushMessages(tx *gorm.DB, msgs []PushMessage) error {
	var userIDs []int32
	for _, msg := range msgs {
		userIDs = append(userIDs, msg.UserID)
	}

	var pushSettings []PushSettings
	err := tx.Model(&PushSettings{}).Where("user_id in (?)", userIDs).
		Find(&pushSettings).Error
	if err != nil {
		log.Printf("read push settings failed: %s", err)
		return err
	}

	pushSettingsMap := make(map[int32]PushSettings)
	for _, s := range pushSettings {
		pushSettingsMap[s.UserID] = s
	}

	for i, msg := range msgs {
		s, ok := pushSettingsMap[msg.UserID]
		if ok {
			if (s.Settings & msg.Type) > 0 {
				msgs[i].DoPush = true
			} else {
				msgs[i].DoPush = false
			}
		} else if (msg.Type & (model.SystemMessage | model.ReplyMeComment)) > 0 {
			msgs[i].DoPush = true
		} else {
			msgs[i].DoPush = false
		}
	}
	return nil
}

func SendToPushService(msgs []PushMessage) {
	log.Printf("going to push messages: %v\n", msgs)

	// 预处理，取出用户的 userId 到库里找他的 openId
	pushUserIDs := make([]int32, 0, len(msgs))
	pushMap := make(map[int32]*PushMessage)
	for _, msg := range msgs {
		if msg.DoPush {
			pushUserIDs = append(pushUserIDs, msg.UserID)
			pushMap[msg.UserID] = &msg
		}
	}

	var users []User
	err := GetDb(false).Model(&User{}).Where("ID in (?)", pushUserIDs).
		Find(&users).Error
	if err != nil {
		log.Printf("read push user infos failed: %s", err)
		return
	}

	// 重构消息表
	pushMessages := make([]PushMessage, 0, len(pushMap))
	for _, user := range users {
		msg := pushMap[user.ID]
		if msg == nil {
			continue
		}
		msg.OpenID = user.WechatOpenId
		pushMessages = append(pushMessages, *msg)
	}

	log.Printf("push messages prepared: %v\n", pushMessages)

	postBody, _ := json.Marshal(pushMessages)
	bytesBody := bytes.NewBuffer(postBody)
	req, err2 := http.NewRequest("POST",
		"http://172.22.0.1:33123/v1/send_messages", bytesBody)
	if err2 != nil {
		log.Printf("push request build failed: %s\n", err2)
		return
	}
	clientHttp := &http.Client{}
	resp, err3 := clientHttp.Do(req)
	if err3 != nil {
		log.Printf("push failed: %s\n", err3)
		return
	}
	_ = resp.Body.Close()
}

func SendDeletionToPushService(commentID int32) {
	postBody, _ := json.Marshal(commentID)
	bytesBody := bytes.NewBuffer(postBody)
	req, err2 := http.NewRequest("POST",
		"http://"+viper.GetString("push_internal_api_listen_address")+"/delete_messages", bytesBody)
	if err2 != nil {
		log.Printf("push request build failed: %s\n", err2)
		return
	}
	clientHttp := &http.Client{}
	resp, err3 := clientHttp.Do(req)
	if err3 != nil {
		log.Printf("push failed: %s\n", err3)
		return
	}
	_ = resp.Body.Close()
}
