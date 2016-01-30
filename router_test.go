package apifaker

import (
	. "github.com/Focinfi/gtester"
	"testing"
)

var testFaker = &ApiFaker{}

func TestNewRouter(t *testing.T) {
	router, err := NewRouterWithPath(testDir+"/users.json", testFaker)
	if err != nil {
		t.Error(err)
	}

	AssertEqual(t, router.Model.Name, "users")
	AssertEqual(t, router.filePath, testDir+"/users.json")
}
