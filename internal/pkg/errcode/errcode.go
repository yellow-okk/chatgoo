package errcode

// ErrCode represents an application-level error code.
type ErrCode int

const (
	Success           ErrCode = 0
	InvalidParams     ErrCode = 40001
	Unauthorized      ErrCode = 40101
	Forbidden         ErrCode = 40301
	NotFound          ErrCode = 40401
	Conflict          ErrCode = 40901
	InternalError     ErrCode = 50001
	UsernameExists    ErrCode = 41001
	InvalidCredential ErrCode = 41002
	SessionNotFound   ErrCode = 41003
	NotParticipant    ErrCode = 41004
)
