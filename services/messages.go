package services

import (
	"context"
	"fmt"

	"github.com/trycourier/courier-go/v4"
	"github.com/trycourier/courier-go/v4/shared"
)

type MessageService struct {
	courierClient courier.Client
}

func NewMessageService(courierClient *courier.Client) *MessageService {
	return &MessageService{
		courierClient: *courierClient,
	}
}

func (ms *MessageService) SendMessage(id uint, name string, price float64) error {
	fmt.Println("sedning message")

	response, err := ms.courierClient.Send.Message(context.TODO(), courier.SendMessageParams{
		Message: courier.SendMessageParamsMessage{
			To: courier.SendMessageParamsMessageToUnion{
				OfUserRecipient: &shared.UserRecipientParam{
					UserID: courier.String("your_user_id"),
				},
			},
			Template: courier.String("your_template_id"),
			Data: map[string]any{
				"foo": "bar",
			},
		},
	})

	fmt.Printf("%v", response)
	return err
}
