package dispatcher

import "github.com/stvp/pager"

var pd *pager.Pager

// SendPush dispatches a message through Pager Duty
func SendPush(message string, details map[string]interface{}) {
	pd.TriggerWithDetails(message, details)
}

// InitPagerDutyInterface initializes pager duty interface
func InitPagerDutyInterface(key string) {
	pd = pager.New(key)
}

func init() {}
