package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ExpediaGroup/flyte-slack/client"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"strings"
)

var (
	chatUpdateSuccessEventDef = flyte.EventDef{Name: chatUpdateSuccessName}
	chatUpdateFailEventDef    = flyte.EventDef{Name: chatUpdateFailName}
)

type ChatUpdateInput struct {
	Text             string `json:"text"`      // Text is what updated message will look like
	MessageTimestamp string `json:"messageTs"` // MessageTimestamp of a message to be updated
	ChannelId        string `json:"channelId"`
}

func (cui *ChatUpdateInput) Validate() error {
	errs := make([]string, 0, 3)

	if cui.Text == "" {
		errs = append(errs, "missing text field")
	}
	if cui.ChannelId == "" {
		errs = append(errs, "missing channel id field")
	}
	if cui.MessageTimestamp == "" {
		errs = append(errs, "missing message timestamp field")
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.New(strings.Join(errs, ", "))
}

type ChatUpdateSuccess struct {
	ChatUpdateInput
}

type ChatUpdateFail struct {
	ChatUpdateInput
	Reason string `json:"reason"`
}

func ChatUpdate(slack client.Slack) flyte.Command {
	return flyte.Command{
		Name:         chatUpdateCommandName,
		OutputEvents: []flyte.EventDef{messageSentEventDef, sendMessageFailedEventDef},
		Handler:      chatUpdateHandler(slack),
	}
}

func chatUpdateHandler(slack client.Slack) func(json.RawMessage) flyte.Event {
	return func(rawInput json.RawMessage) flyte.Event {
		input := ChatUpdateInput{}
		if err := json.Unmarshal(rawInput, &input); err != nil {
			return flyte.NewFatalEvent(fmt.Sprintf("input is not valid: %v", err))
		}

		err := input.Validate()
		if err != nil {
			return chatUpdateFailEvent(input, err.Error())
		}

		err = slack.UpdateMessage(input.ChannelId, input.MessageTimestamp, input.Text)
		if err != nil {
			return chatUpdateFailEvent(input, err.Error())
		}

		return chatUpdateSuccessEvent(input)
	}
}

func chatUpdateSuccessEvent(input ChatUpdateInput) flyte.Event {
	return flyte.Event{
		EventDef: chatUpdateSuccessEventDef,
		Payload: ChatUpdateSuccess{
			ChatUpdateInput: input,
		},
	}
}

func chatUpdateFailEvent(input ChatUpdateInput, reason string) flyte.Event {
	return flyte.Event{
		EventDef: chatUpdateFailEventDef,
		Payload: ChatUpdateFail{
			ChatUpdateInput: input,
			Reason:          reason,
		},
	}
}
