package demorecord

import (
	"errors"
	"fmt"
)

type DemoRecordEventType byte

const (
	PlayerJoin       DemoRecordEventType = 1
	PlayerDisconnect DemoRecordEventType = 2
	PlayerPerk       DemoRecordEventType = 3
	PlayerDied       DemoRecordEventType = 4

	GlobalWaveStart DemoRecordEventType = 17
	GlobalWaveEnd   DemoRecordEventType = 18
	GlobalZedTime   DemoRecordEventType = 19
	GlobalZedsLeft  DemoRecordEventType = 20

	EventKill     DemoRecordEventType = 33
	EventBuffs    DemoRecordEventType = 34
	EventHpChange DemoRecordEventType = 35
	EventHuskRage DemoRecordEventType = 36
)

type DemoRecordHeader struct {
	Header    string `json:"header"`
	Version   byte   `json:"version"`
	SessionId int    `json:"session_id"`
}

type DemoRecordRawEvent struct {
	Tick int            `json:"tick"`
	Type byte           `json:"event_type"`
	Data map[string]any `json:"payload,omitempty"`
}

type DemoRecordRaw struct {
	Header *DemoRecordHeader     `json:"header"`
	Events []*DemoRecordRawEvent `json:"events"`
}

type unexpectedTokenError struct {
	Pos      int
	Expected string
	Actual   string
}

func (e *unexpectedTokenError) Error() string {
	return fmt.Sprintf("unexpected token on pos %v, expected %v, got %v", e.Pos, e.Expected, e.Actual)
}

type unexpectedEventSizeError struct {
	Expected int
	Actual   int
}

func (e *unexpectedEventSizeError) Error() string {
	return fmt.Sprintf("unexpected event size, expected at least %v, got %v", e.Expected, e.Actual)
}

func Parse(raw []byte) (*DemoRecordRaw, error) {
	demo := DemoRecordRaw{}

	header, err := parseHeader(raw)
	if err != nil {
		return nil, err
	}
	demo.Header = header

	start := 12
	for pos := start; pos < len(raw); {
		event, size, err := parseEvent(raw[pos:])
		if err != nil {
			return nil, err
		}

		demo.Events = append(demo.Events, event)
		pos += size
	}

	return &demo, nil
}

func parseHeader(raw []byte) (*DemoRecordHeader, error) {
	if len(raw) < 11 {
		return nil, errors.New(fmt.Sprintf("unexpected header size, expected 11, got %v", len(raw)))
	}

	header := DemoRecordHeader{
		Header:    readString(raw, 0, 6),
		Version:   readByte(raw, 6),
		SessionId: readInt(raw, 7),
	}

	if header.Header != "kf2rec" {
		return nil, errors.New(
			fmt.Sprintf("unexpected header.Header: expected: %v, got: %v", "kf2rec", header.Header),
		)
	}

	if raw[11] != 0 {
		return nil, &unexpectedTokenError{Pos: 11, Expected: "\\0", Actual: string(raw[11])}
	}

	return &header, nil
}

func parseEvent(raw []byte) (*DemoRecordRawEvent, int, error) {
	if len(raw) < 6 {
		return nil, 0, errors.New(fmt.Sprintf("unexpected event size, expected at least 6, got %v", len(raw)))
	}

	event := DemoRecordRawEvent{
		Tick: readInt(raw, 0),
		Type: readByte(raw, 4),
	}

	var parseEventPayloadFunc func(byte []byte) (map[string]any, int, error)

	if event.Type == byte(PlayerJoin) {
		parseEventPayloadFunc = parsePlayerJoinedEvent
	} else if event.Type == byte(PlayerDisconnect) {
		parseEventPayloadFunc = parsePlayerDisconnectedEvent
	} else if event.Type == byte(PlayerPerk) {
		parseEventPayloadFunc = parsePlayerPerkEvent
	} else if event.Type == byte(PlayerDied) {
		parseEventPayloadFunc = parsePlayerDiedEvent
	} else if event.Type == byte(GlobalWaveStart) {
		parseEventPayloadFunc = parseWaveStartEvent
	} else if event.Type == byte(GlobalZedsLeft) {
		parseEventPayloadFunc = parseZedsLeftEvent
	} else if event.Type == byte(EventKill) {
		parseEventPayloadFunc = parseKillEvent
	} else if event.Type == byte(EventBuffs) {
		parseEventPayloadFunc = parseBuffsEvent
	} else if event.Type == byte(EventHpChange) {
		parseEventPayloadFunc = parseHpChangeEvent
	} else if event.Type == byte(EventHuskRage) {
		parseEventPayloadFunc = parseHuskRageEvent
	} else {
		parseEventPayloadFunc = nil
	}

	if parseEventPayloadFunc != nil {
		data, size, err := parseEventPayloadFunc(raw[5:])
		if err != nil {
			return nil, 0, err
		}

		event.Data = data

		// null terminator calculates inside payload func
		return &event, 5 + size, nil
	}

	return &event, 6, nil
}

func parsePlayerJoinedEvent(raw []byte) (map[string]any, int, error) {
	if len(raw) < 2 {
		return nil, 0, &unexpectedEventSizeError{Expected: 2, Actual: len(raw)}
	}

	values := map[string]any{}

	values["user_id"] = readByte(raw, 0)
	values["user_type"] = readByte(raw, 1)

	if values["user_type"].(byte) == 1 {
		if len(raw) < 20 {
			return nil, 0, &unexpectedEventSizeError{Expected: 20, Actual: len(raw)}
		}

		values["unique_id"] = readString(raw, 2, 17)

		if raw[19] != 0 {
			return nil, 0, &unexpectedTokenError{Pos: 19, Expected: "\\0", Actual: string(raw[19])}
		}

		return values, 20, nil
	} else {
		if len(raw) < 21 {
			return nil, 0, &unexpectedEventSizeError{Expected: 21, Actual: len(raw)}
		}

		values["unique_id"] = readString(raw, 2, 18)

		if raw[20] != 0 {
			return nil, 0, &unexpectedTokenError{Pos: 20, Expected: "\\0", Actual: string(raw[20])}
		}

		return values, 21, nil
	}
}

func parsePlayerDisconnectedEvent(raw []byte) (map[string]any, int, error) {
	if len(raw) < 2 {
		return nil, 0, &unexpectedEventSizeError{Expected: 2, Actual: len(raw)}
	}

	values := map[string]any{}

	values["user_id"] = readByte(raw, 0)

	if raw[1] != 0 {
		return nil, 0, &unexpectedTokenError{Pos: 1, Expected: "\\0", Actual: string(raw[1])}
	}

	return values, 2, nil
}

func parsePlayerPerkEvent(raw []byte) (map[string]any, int, error) {
	if len(raw) < 3 {
		return nil, 0, &unexpectedEventSizeError{Expected: 3, Actual: len(raw)}
	}

	values := map[string]any{}

	values["user_id"] = readByte(raw, 0)
	values["perk"] = readByte(raw, 1)

	if raw[2] != 0 {
		return nil, 0, &unexpectedTokenError{Pos: 2, Expected: "\\0", Actual: string(raw[2])}
	}

	return values, 3, nil
}

func parseWaveStartEvent(raw []byte) (map[string]any, int, error) {
	if len(raw) < 6 {
		return nil, 0, &unexpectedEventSizeError{Expected: 6, Actual: len(raw)}
	}

	values := map[string]any{}

	values["wave"] = readByte(raw, 0)
	values["zeds_left"] = readInt(raw, 1)

	if raw[5] != 0 {
		return nil, 0, &unexpectedTokenError{Pos: 5, Expected: "\\0", Actual: string(raw[5])}
	}

	return values, 6, nil
}

func parseZedsLeftEvent(raw []byte) (map[string]any, int, error) {
	if len(raw) < 5 {
		return nil, 0, &unexpectedEventSizeError{Expected: 5, Actual: len(raw)}
	}

	values := map[string]any{}

	values["zeds_left"] = readInt(raw, 0)

	if raw[4] != 0 {
		return nil, 0, &unexpectedTokenError{Pos: 4, Expected: "\\0", Actual: string(raw[4])}
	}

	return values, 5, nil
}

func parseKillEvent(raw []byte) (map[string]any, int, error) {
	if len(raw) < 3 {
		return nil, 0, &unexpectedEventSizeError{Expected: 3, Actual: len(raw)}
	}

	values := map[string]any{}

	values["user_id"] = readByte(raw, 0)
	values["zed"] = readByte(raw, 1)

	if raw[2] != 0 {
		return nil, 0, &unexpectedTokenError{Pos: 2, Expected: "\\0", Actual: string(raw[2])}
	}

	return values, 3, nil
}

func parseBuffsEvent(raw []byte) (map[string]any, int, error) {
	if len(raw) < 3 {
		return nil, 0, &unexpectedEventSizeError{Expected: 3, Actual: len(raw)}
	}

	values := map[string]any{}

	values["user_id"] = readByte(raw, 0)
	values["max_buffs"] = readByte(raw, 1)

	if raw[2] != 0 {
		return nil, 0, &unexpectedTokenError{Pos: 2, Expected: "\\0", Actual: string(raw[2])}
	}

	return values, 3, nil
}

func parseHpChangeEvent(raw []byte) (map[string]any, int, error) {
	if len(raw) < 7 {
		return nil, 0, &unexpectedEventSizeError{Expected: 7, Actual: len(raw)}
	}

	values := map[string]any{}

	values["user_id"] = readByte(raw, 0)
	values["health"] = readInt(raw, 1)
	values["armor"] = readByte(raw, 5)

	if raw[6] != 0 {
		return nil, 0, &unexpectedTokenError{Pos: 6, Expected: "\\0", Actual: string(raw[6])}
	}

	return values, 7, nil
}

func parseHuskRageEvent(raw []byte) (map[string]any, int, error) {
	if len(raw) < 2 {
		return nil, 0, &unexpectedEventSizeError{Expected: 2, Actual: len(raw)}
	}

	values := map[string]any{}

	values["user_id"] = readByte(raw, 0)

	if raw[1] != 0 {
		return nil, 0, &unexpectedTokenError{Pos: 1, Expected: "\\0", Actual: string(raw[1])}
	}

	return values, 2, nil
}

func parsePlayerDiedEvent(raw []byte) (map[string]any, int, error) {
	if len(raw) < 3 {
		return nil, 0, &unexpectedEventSizeError{Expected: 3, Actual: len(raw)}
	}

	values := map[string]any{}

	values["user_id"] = readByte(raw, 0)
	values["cause"] = readByte(raw, 1)

	if raw[2] != 0 {
		return nil, 0, &unexpectedTokenError{Pos: 2, Expected: "\\0", Actual: string(raw[2])}
	}

	return values, 3, nil
}
