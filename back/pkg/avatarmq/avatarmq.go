package avatarmq

const (
	avatarEventTopic      = "avatar_events"
	eventTypeAvatarSubmit = "AVATAR_SUBMIT"
)

type AvatarModerationPayload struct {
	UserID           uint64 `json:"user_id"`
	AvatarURL        string `json:"avatar_url"`
	DefaultAvatarURL string `json:"default_avatar_url"`
	RequestID        string `json:"request_id"`
	SubmittedAt      int64  `json:"submitted_at"`
}

func AvatarEventTopic() string {
	return avatarEventTopic
}

func EventTypeAvatarSubmit() string {
	return eventTypeAvatarSubmit
}
