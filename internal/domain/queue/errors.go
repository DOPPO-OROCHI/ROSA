package queue

import "errors"

//ошибки в очереди
var ErrNilUser = errors.New("nil user")
var ErrUserAlreadyInQueue = errors.New("user alredy in queue")
var ErrUserQueuePenalty = errors.New("user queue penalty")
var ErrInvalidUserID = errors.New("invalid user id")
var ErrNilQueue = errors.New("nil queue")
var ErrUserNotFoundInQueue = errors.New("user not found in queue")
var ErrBadUserID = errors.New("bad user id")
var ErrBadStatus = errors.New("bad queue status")
var ErrBadPenaltyTime = errors.New("bad penalty time")
