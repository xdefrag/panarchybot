package metrics

const (
	labelStatus        = "status"
	labelStatusSuccess = "success"
	labelStatusError   = "error"
)

func getStatus(err error) string {
	if err != nil {
		return labelStatusError
	}

	return labelStatusSuccess
}
