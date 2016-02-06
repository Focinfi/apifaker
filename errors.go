package apifaker

import (
	"fmt"
)

func JsonFileErrorf(format string, a ...interface{}) error {
	return fmt.Errorf("Error [apifaker-JosnFile]: "+format, a...)
}

func ColumnsErrorf(format string, a ...interface{}) error {
	return fmt.Errorf("Error [apifaker-columns]: "+format, a...)
}

func HasOneErrorf(format string, a ...interface{}) error {
	return fmt.Errorf("Error [apifaker-has_one]: "+format, a...)
}

func HasManyErrorf(format string, a ...interface{}) error {
	return fmt.Errorf("Error [apifaker-has_many]: "+format, a...)
}

func SeedsErrorf(format string, a ...interface{}) error {
	return fmt.Errorf("Error [apifaker-seeds]: "+format, a...)
}

func ResponseErrorMsg(err error) map[string]string {
	return map[string]string{"message": err.Error()}
}
