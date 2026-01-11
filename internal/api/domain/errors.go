package domain

type InvalidPayloadError struct {
	Fields map[string]string
}

func (e InvalidPayloadError) Error() string {
	return "invalid payload"
}
