package ravelerrors

var (
	ErrReservationNotFound  = NewNotFound("reservation not found")
	ErrInstanceNotFound     = NewNotFound("instance not found")
	ErrInstanceIsRunning    = NewFailedPrecondition("instance is running")
	ErrInstanceIsNotRunning = NewFailedPrecondition("instance is not running")
)
