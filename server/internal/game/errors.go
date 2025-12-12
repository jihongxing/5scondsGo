package game

import "errors"

// 观战者相关错误
var (
	ErrAlreadyParticipant    = errors.New("already a participant")
	ErrAlreadySpectator      = errors.New("already a spectator")
	ErrSpectatorLimitReached = errors.New("spectator limit reached")
	ErrNotSpectator          = errors.New("not a spectator")
	ErrRoomFull              = errors.New("room is full")
	ErrNotParticipant        = errors.New("not a participant")
)
