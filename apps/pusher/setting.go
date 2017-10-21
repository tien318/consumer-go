package pusher

import "encoding/json"

type ReminderSetting OtherReminderSetting

func UnmarshalReminderSetting(data []byte) (ReminderSetting, error) {
	var r ReminderSetting
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ReminderSetting) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type OtherReminderSetting struct {
	Buttons   []Button `json:"buttons"`
	Enable    bool     `json:"enable"`
	Message   string   `json:"message"`
	Subject   string   `json:"subject"`
	KeyString string
}

type Button struct {
	Text string `json:"text"`
	Url  string `json:"url"`
}
