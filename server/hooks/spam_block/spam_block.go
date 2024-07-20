package main

import (
	"encoding/json"
	"fmt"
	"github.com/Jinnrry/pmail/dto/parsemail"
	"github.com/Jinnrry/pmail/hooks/framework"
	"github.com/Jinnrry/pmail/hooks/spam_block/tools"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type SpamBlock struct {
	cfg SpamBlockConfig
	hc  *http.Client
}

func (s SpamBlock) SendBefore(ctx *context.Context, email *parsemail.Email) {

}

func (s SpamBlock) SendAfter(ctx *context.Context, email *parsemail.Email, err map[string]error) {

}

func (s SpamBlock) ReceiveParseBefore(ctx *context.Context, email *[]byte) {

}

type ModelResponse struct {
	Predictions [][]float64 `json:"predictions"`
}

type ApiRequest struct {
	Instances []InstanceItem `json:"instances"`
}

type InstanceItem struct {
	Token []string `json:"token"`
}

func (s SpamBlock) ReceiveParseAfter(ctx *context.Context, email *parsemail.Email) {

	reqData := ApiRequest{
		Instances: []InstanceItem{
			{
				Token: []string{
					fmt.Sprintf("%s %s", email.Subject, tools.Trim(tools.TrimHtml(string(email.HTML)))),
				},
			},
		},
	}

	str, _ := json.Marshal(reqData)

	resp, err := s.hc.Post(s.cfg.ApiURL, "application/json", strings.NewReader(string(str)))
	if err != nil {
		log.Errorf("API Error: %v", err)
		return
	}

	body, _ := io.ReadAll(resp.Body)

	modelResponse := ModelResponse{}
	err = json.Unmarshal(body, &modelResponse)
	if err != nil {
		log.WithContext(ctx).Errorf("API Error: %v", err)
		return
	}

	if len(modelResponse.Predictions) == 0 {
		log.WithContext(ctx).Errorf("API Response Error: %v", string(body))
		return
	}

	classes := modelResponse.Predictions[0]
	if len(classes) != 3 {
		return
	}
	var maxScore float64
	var maxClass int
	for i, score := range classes {
		if score > maxScore {
			maxScore = score
			maxClass = i
		}
	}

	switch maxClass {
	case 0:
		log.WithContext(ctx).Infof("[Spam Check Result: Normal] %s", email.Subject)
	case 1:
		log.WithContext(ctx).Infof("[Spam Check Result: Spam ] %s", email.Subject)
	case 2:
		log.WithContext(ctx).Infof("[Spam Check Result: Blackmail ] %s", email.Subject)
	}

	if maxClass != 0 {
		email.Status = 3
	}
}

func (s SpamBlock) ReceiveSaveAfter(ctx *context.Context, email *parsemail.Email, ue []*models.UserEmail) {

}

type SpamBlockConfig struct {
	ApiURL     string `json:"apiURL"`
	ApiTimeout int    `json:"apiTimeout"` // 单位毫秒
}

func NewSpamBlockHook() *SpamBlock {

	var pluginConfig SpamBlockConfig
	if _, err := os.Stat("./plugins/spam_block_config.json"); err == nil {
		cfgData, err := os.ReadFile("./plugins/spam_block_config.json")
		if err == nil {
			json.Unmarshal(cfgData, &pluginConfig)
		}
	} else {
		log.Infof("No Config file found")
		return nil
	}

	log.Infof("Config: %+v", pluginConfig)
	if pluginConfig.ApiURL == "" {
		pluginConfig.ApiURL = "http://localhost:8501/v1/models/emotion_model:predict"
	}

	if pluginConfig.ApiTimeout == 0 {
		pluginConfig.ApiTimeout = 3000
	}

	hc := &http.Client{
		Timeout: time.Duration(pluginConfig.ApiTimeout) * time.Millisecond,
	}

	return &SpamBlock{
		cfg: pluginConfig,
		hc:  hc,
	}
}

func main() {
	log.Infof("SpamBlockPlug Star Success")
	instance := NewSpamBlockHook()
	if instance == nil {
		return
	}
	framework.CreatePlugin("spam_block", instance).Run()
}
