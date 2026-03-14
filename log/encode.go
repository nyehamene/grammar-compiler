package log

func EncodeError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
