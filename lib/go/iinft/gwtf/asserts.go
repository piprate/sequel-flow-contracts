package gwtf

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TransactionResult struct {
	Err     error
	Events  []*FormatedEvent
	Testing *testing.T
}

func (tb FlowTransactionBuilder) Test(t *testing.T) TransactionResult {
	locale, _ := time.LoadLocation("UTC")
	time.Local = locale
	events, err := tb.RunE()
	formattedEvents := make([]*FormatedEvent, len(events))
	for i, event := range events {
		ev := ParseEvent(event, uint64(0), time.Unix(0, 0), []string{})
		formattedEvents[i] = ev
	}
	return TransactionResult{
		Err:     err,
		Events:  formattedEvents,
		Testing: t,
	}
}

func (t TransactionResult) AssertFailure(msg string) TransactionResult {
	assert.Error(t.Testing, t.Err)
	if t.Err != nil {
		assert.Contains(t.Testing, t.Err.Error(), msg)
	}
	return t
}

func (t TransactionResult) AssertSuccess() TransactionResult {
	assert.NoError(t.Testing, t.Err)
	return t
}

func (t TransactionResult) AssertEventCount(number int) TransactionResult {
	assert.Equal(t.Testing, number, len(t.Events))
	return t
}

func (t TransactionResult) AssertNoEvents() TransactionResult {
	assert.Empty(t.Testing, t.Events)

	for _, ev := range t.Events {
		t.Testing.Log(ev.String())
	}

	return t
}

func (t TransactionResult) AssertEmitEventName(event ...string) TransactionResult {
	eventNames := make([]string, len(t.Events))
	for i, fe := range t.Events {
		eventNames[i] = fe.Name
	}

	for _, ev := range event {
		assert.Contains(t.Testing, eventNames, ev)
	}

	for _, ev := range t.Events {
		t.Testing.Log(ev.String())
	}

	return t
}

func (t TransactionResult) AssertEmitEventJSON(event ...string) TransactionResult {

	jsonEvents := make([]string, len(t.Events))
	for i, fe := range t.Events {
		jsonEvents[i] = fe.String()
	}

	for _, ev := range event {
		assert.Contains(t.Testing, jsonEvents, ev)
	}

	for _, ev := range t.Events {
		t.Testing.Log(ev.String())
	}

	return t
}

func (t TransactionResult) AssertPartialEvent(expected *FormatedEvent) TransactionResult {

	events := t.Events
	for index, ev := range events {
		// TODO do we need more then just name here?
		if ev.Name == expected.Name {
			for key := range ev.Fields {
				_, exist := expected.Fields[key]
				if !exist {
					delete(events[index].Fields, key)
				}
			}
		}
	}

	assert.Contains(t.Testing, events, expected)

	for _, ev := range events {
		t.Testing.Log(ev.String())
	}

	return t
}
func (t TransactionResult) AssertEmitEvent(event ...*FormatedEvent) TransactionResult {
	for _, ev := range event {
		assert.Contains(t.Testing, t.Events, ev)
	}

	for _, ev := range t.Events {
		t.Testing.Log(ev.String())
	}

	return t
}

func (t TransactionResult) AssertDebugLog(message ...string) TransactionResult {
	var logMessages []interface{}
	for _, fe := range t.Events {
		if strings.HasSuffix(fe.Name, "Debug.Log") {
			logMessages = append(logMessages, fe.Fields["msg"])
		}
	}

	for _, ev := range message {
		assert.Contains(t.Testing, logMessages, ev)
	}
	return t
}
