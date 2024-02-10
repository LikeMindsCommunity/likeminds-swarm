package worker

// Task Names for each task type
const (
	TaskSendDeleteTopicsFromPosts = "task:DeleteTopicsFromPosts"
)

// PayloadSendDeleteTopicsFromPostsTask | Payload for the task to delete topics from posts
type PayloadSendDeleteTopicsFromPostsTask struct {
	TopicIds []string `json:"topic_ids"`
}
