package apifaker

import (
	"fmt"
)

type err struct {
	prefix  string
	content string
}

func (e err) Error() string {
	return fmt.Sprintf("%s %s", e.prefix, e.content)
}

func JsonFileErrorf(format string, a ...interface{}) err {
	fmt.Sprintf(format, a...)
	return err{prefix: "Error [apifaker-JosnFile]:", content: fmt.Sprintf(format, a...)}
}

func ColumnsErrorf(format string, a ...interface{}) err {
	return err{prefix: "Error [apifaker-columns]:", content: fmt.Sprintf(format, a...)}
}

func HasOneErrorf(format string, a ...interface{}) err {
	return err{prefix: "Error [apifaker-has_one]:", content: fmt.Sprintf(format, a...)}
}

func HasManyErrorf(format string, a ...interface{}) err {
	return err{prefix: "Error [apifaker-has_many]:", content: fmt.Sprintf(format, a...)}
}

func SeedsErrorf(format string, a ...interface{}) err {
	return err{prefix: "Error [apifaker-seeds]:", content: fmt.Sprintf(format, a...)}
}

func ResponseErrorMsg(err error) map[string]string {
	return map[string]string{"message": err.Error()}
}
