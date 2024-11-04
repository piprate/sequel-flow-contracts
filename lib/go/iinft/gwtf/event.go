package gwtf

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
)

// EventFetcherBuilder builder to hold info about eventhook context.
type EventFetcherBuilder struct {
	GoWithTheFlow         *GoWithTheFlow
	EventsAndIgnoreFields map[string][]string
	FromIndex             int64
	EndAtCurrentHeight    bool
	EndIndex              uint64
	ProgressFile          string
	NumberOfWorkers       int
	EventBatchSize        uint64
}

// EventFetcher create an event fetcher builder.
func (f *GoWithTheFlow) EventFetcher() EventFetcherBuilder {
	return EventFetcherBuilder{
		GoWithTheFlow:         f,
		EventsAndIgnoreFields: map[string][]string{},
		EndAtCurrentHeight:    true,
		FromIndex:             -10,
		ProgressFile:          "",
		EventBatchSize:        250,
		NumberOfWorkers:       20,
	}
}

// Workers sets the number of workers.
func (e EventFetcherBuilder) Workers(workers int) EventFetcherBuilder {
	e.NumberOfWorkers = workers
	return e
}

// BatchSize sets the size of a batch
func (e EventFetcherBuilder) BatchSize(batchSize uint64) EventFetcherBuilder {
	e.EventBatchSize = batchSize
	return e
}

// Event fetches and Events and all its fields
func (e EventFetcherBuilder) Event(eventName string) EventFetcherBuilder {
	e.EventsAndIgnoreFields[eventName] = []string{}
	return e
}

// EventIgnoringFields fetch event and ignore the specified fields
func (e EventFetcherBuilder) EventIgnoringFields(eventName string, ignoreFields []string) EventFetcherBuilder {
	e.EventsAndIgnoreFields[eventName] = ignoreFields
	return e
}

// Start specify what blockHeight to fetch starting atm. This can be negative related to end/until
func (e EventFetcherBuilder) Start(blockHeight int64) EventFetcherBuilder {
	e.FromIndex = blockHeight
	return e
}

// From specify what blockHeight to fetch from. This can be negative related to end.
func (e EventFetcherBuilder) From(blockHeight int64) EventFetcherBuilder {
	e.FromIndex = blockHeight
	return e
}

// End specify what index to end at
func (e EventFetcherBuilder) End(blockHeight uint64) EventFetcherBuilder {
	e.EndIndex = blockHeight
	e.EndAtCurrentHeight = false
	return e
}

// Last fetch events from the number last blocks
func (e EventFetcherBuilder) Last(number uint64) EventFetcherBuilder {
	e.EndAtCurrentHeight = true
	e.FromIndex = -int64(number) //nolint:gosec
	return e
}

// Until specify what index to end at
func (e EventFetcherBuilder) Until(blockHeight uint64) EventFetcherBuilder {
	e.EndIndex = blockHeight
	e.EndAtCurrentHeight = false
	return e
}

// UntilCurrent Specify to fetch events until the current Block
func (e EventFetcherBuilder) UntilCurrent() EventFetcherBuilder {
	e.EndAtCurrentHeight = true
	e.EndIndex = 0
	return e
}

// TrackProgressIn Specify a file to store progress in
func (e EventFetcherBuilder) TrackProgressIn(fileName string) EventFetcherBuilder {
	e.ProgressFile = fileName
	e.EndIndex = 0
	e.FromIndex = 0
	e.EndAtCurrentHeight = true
	return e
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func writeProgressToFile(fileName string, blockHeight uint64) error {

	err := os.WriteFile(fileName, []byte(fmt.Sprintf("%d", blockHeight)), 0o0644) //nolint:gosec // inherited from GWTF

	if err != nil {
		return fmt.Errorf("could not create initial progress file %w", err)
	}
	return nil
}

func readProgressFromFile(fileName string) (int64, error) {
	dat, err := os.ReadFile(fileName)
	if err != nil {
		return 0, fmt.Errorf("ProgressFile is not valid %w", err)
	}

	stringValue := strings.TrimSpace(string(dat))

	return strconv.ParseInt(stringValue, 10, 64)
}

// Run runs the eventfetcher returning events or an error
func (e EventFetcherBuilder) Run(ctx context.Context) ([]*FormatedEvent, error) {

	// if we have a progress file read the value from it and set it as oldHeight
	if e.ProgressFile != "" {

		present, err := exists(e.ProgressFile)
		if err != nil {
			return nil, err
		}

		if !present {
			err := writeProgressToFile(e.ProgressFile, 0)
			if err != nil {
				return nil, fmt.Errorf("could not create initial progress file %w", err)
			}

			e.FromIndex = 0
		} else {
			oldHeight, err := readProgressFromFile(e.ProgressFile)
			if err != nil {
				return nil, fmt.Errorf("could not parse progress file as block height %w", err)
			}
			e.FromIndex = oldHeight
		}
	}

	endIndex := e.EndIndex
	if e.EndAtCurrentHeight {
		block, err := e.GoWithTheFlow.Services.GetBlock(ctx, flowkit.LatestBlockQuery)
		if err != nil {
			return nil, err
		}
		endIndex = block.Height
	}

	fromIndex := e.FromIndex
	// if we have a negative fromIndex is relative to endIndex
	if e.FromIndex <= 0 {
		fromIndex = int64(endIndex) + e.FromIndex //nolint:gosec
	}

	if fromIndex < 0 {
		return nil, fmt.Errorf("FromIndex is negative")
	}

	e.GoWithTheFlow.Logger.Info(fmt.Sprintf("Fetching events from %d to %d", fromIndex, endIndex))

	events := make([]string, len(e.EventsAndIgnoreFields))
	for key := range e.EventsAndIgnoreFields {
		events = append(events, key)
	}

	blockEvents, err := e.GoWithTheFlow.Services.GetEvents(ctx, events, uint64(fromIndex), endIndex, &flowkit.EventWorker{
		Count:           e.NumberOfWorkers,
		BlocksPerWorker: e.EventBatchSize,
	})
	if err != nil {
		return nil, err
	}

	formatedEvents := FormatEvents(blockEvents, e.EventsAndIgnoreFields)

	if e.ProgressFile != "" {
		err := writeProgressToFile(e.ProgressFile, endIndex+1)
		if err != nil {
			return nil, fmt.Errorf("could not write progress to file %w", err)
		}
	}
	sort.Slice(formatedEvents, func(i, j int) bool {
		return formatedEvents[i].BlockHeight < formatedEvents[j].BlockHeight
	})

	return formatedEvents, nil
}

// PrintEvents prints th events, ignoring fields specified for the given event typeID
func PrintEvents(events []flow.Event, ignoreFields map[string][]string) {
	if len(events) > 0 {
		log.Println("EVENTS")
		log.Println("======")
	}

	for _, event := range events {
		//TODO: does this change work on mainnet/testnet?
		ignoreFieldsForType := ignoreFields[event.Type]
		ev := ParseEvent(event, uint64(0), time.Now(), ignoreFieldsForType)
		prettyJSON, err := json.MarshalIndent(ev, "", "    ")

		if err != nil {
			panic(err)
		}

		log.Printf("%s\n", string(prettyJSON))
	}
	if len(events) > 0 {
		log.Println("======")
	}
}

// FormatEvents
func FormatEvents(blockEvents []flow.BlockEvents, ignoreFields map[string][]string) []*FormatedEvent {
	var events []*FormatedEvent

	for _, blockEvent := range blockEvents {
		for _, event := range blockEvent.Events {
			ev := ParseEvent(event, blockEvent.Height, blockEvent.BlockTimestamp, ignoreFields[event.Type])
			events = append(events, ev)
		}
	}
	return events
}

// ParseEvent parses a flow event into a more terse representation
func ParseEvent(event flow.Event, blockHeight uint64, time time.Time, ignoreFields []string) *FormatedEvent {

	finalFields := map[string]interface{}{}

	for name, field := range event.Value.FieldsMappedByName() {

		skip := false

		for _, ignoreField := range ignoreFields {
			if ignoreField == name {
				skip = true
			}
		}
		if skip {
			continue
		}

		finalFields[name] = CadenceValueToInterface(field)
	}
	return &FormatedEvent{
		Name:        event.Type,
		Fields:      finalFields,
		BlockHeight: blockHeight,
		Time:        time,
	}
}

// FormatedEvent event in a more condensed formated form
type FormatedEvent struct {
	Name        string                 `json:"name"`
	BlockHeight uint64                 `json:"blockHeight,omitempty"`
	Time        time.Time              `json:"time,omitempty"`
	Fields      map[string]interface{} `json:"fields"`
}

func NewTestEvent(name string, fields map[string]interface{}) *FormatedEvent {
	loc, _ := time.LoadLocation("UTC")
	// handle err
	time.Local = loc // -> this is setting the global timezone
	return &FormatedEvent{
		Name:        name,
		BlockHeight: 0,
		Time:        time.Unix(0, 0),
		Fields:      fields,
	}
}

// String pretty print an event as a String
func (e FormatedEvent) String() string {
	j, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(j)
}
