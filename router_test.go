package apifaker

import (
	"github.com/Focinfi/gtester"
	"testing"
)

func TestNewRouter(t *testing.T) {
	router, err := NewRouterWithPath(testDir + "/users.json")
	if err != nil {
		t.Error(err)
	}

	gtester.AssertEqual(t, router.Name, "users")
	gtester.AssertEqual(t, router.filePath, testDir+"/users.json")
}
