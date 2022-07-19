package battle

func CalcNextQueueStatus(length int64) QueueStatusResponse {
	return QueueStatusResponse{
		QueueLength: length, // return the current queue length
	}
}
