package shared

type TaskType string

const (
	Update     TaskType = "update"
	Compliance TaskType = "compliance"
)

func Tables(task string, tt TaskType) []string {
	return []string{
		string(tt) + "_" + task + "_admin",
		string(tt) + "_" + task + "_exclusions",
		string(tt) + "_" + task + "_tracking",
		string(tt) + "_" + task + "_task",
		string(tt) + "_" + task + "_updates",
	}
}
