package main

import (
	"fmt"
	"sort"
)

var (
	ErrDateNotInSequence              = fmt.Errorf("event validator: ToDate is before FromDate")
	FmtErrTwoSessionsSameTimeSameRoom = "event validator: session %s overrides the previous session in same room %s"
	FmtErrSessionDateNotInSequence    = "event validator: session %s ToDate is before FromDate"
	FmtErrSessionRoomNotInEvent       = "event validator: session %s has defined room not defined in event"
	FmtErrSessionSpeakerNotInEvent    = "event validator: session %s has defined speak %s not defined in event"
)

func ValidateEvent(e *Event) error {

	if e.FromDate >= e.ToDate {
		return ErrDateNotInSequence
	}

	// Prepare map to validate rooms
	timeMap := make(map[string]int64)
	for _, room := range e.Rooms {
		timeMap[room.Name] = -100
	}

	// Prepare speakers to validate
	speakerMap := make(map[string]bool)
	for _, spkr := range e.Speakers {
		speakerMap[spkr] = true
	}

	sort.Sort(Sessions(e.Sessions))

	for _, session := range e.Sessions {
		for _, spkr := range session.Speaker {
			if !speakerMap[spkr] {
				return fmt.Errorf(FmtErrSessionSpeakerNotInEvent,
					session.SessionToken,
					spkr)
			}
		}
		if timeMap[session.Room] == 0 {
			return fmt.Errorf(FmtErrSessionRoomNotInEvent,
				session.SessionToken)
		}
		if session.From < timeMap[session.Room] {
			return fmt.Errorf(FmtErrTwoSessionsSameTimeSameRoom,
				session.SessionToken,
				session.Room)
		}
		if session.From >= session.To {
			return fmt.Errorf(FmtErrSessionDateNotInSequence,
				session.SessionToken)
		}
		timeMap[session.Room] = session.To
	}

	return nil
}
