package logs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SLACKWriter implements beego LoggerInterface and is used to send jiaoliao webhook
type DingtalkWriter struct {
	WebhookURL string `json:"webhookurl"`
	Level      int    `json:"level"`
}

// newSLACKWriter create jiaoliao writer.
func newDingtalkWriter() Logger {
	return &DingtalkWriter{Level: LevelTrace}
}

// Init SLACKWriter with json config string
func (s *DingtalkWriter) Init(jsonconfig string) error {
	return json.Unmarshal([]byte(jsonconfig), s)
}

// WriteMsg write message in smtp writer.
// it will send an email with subject and only this message.
func (s *DingtalkWriter) WriteMsg(when time.Time, msg string, level int) error {
	if level > s.Level {
		return nil
	}
	dingtalk := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("%s %s", when.Format("2006-01-02 15:04:05"), msg),
		},
	}
	text, err := json.Marshal(dingtalk)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(text)
	request, err := http.NewRequest("POST", s.WebhookURL, reader)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Post webhook failed %s %d", resp.Status, resp.StatusCode)
	}
	return nil
}

func (s *DingtalkWriter) WriteOriginalMsg(when time.Time, msg string, level int) error {
	return s.WriteMsg(when, msg, level)
}

// Flush implementing method. empty.
func (s *DingtalkWriter) Flush() {
}

// Destroy implementing method. empty.
func (s *DingtalkWriter) Destroy() {
}

func init() {
	Register(AdapterDingtalk, newDingtalkWriter)
}
