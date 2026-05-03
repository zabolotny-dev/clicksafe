package messageapp

import "github.com/zabolotny-dev/clicksafe/business/domain/messagebus"

var orderByFields = map[string]string{
	"message_id": messagebus.OrderByID,
	"label":      messagebus.OrderByLabel,
	"from_email": messagebus.OrderByEmail,
	"from_name":  messagebus.OrderByName,
	"subject":    messagebus.OrderBySubject,
}
