package autostart

import (
	"fmt"
	"math/rand"
	"testing"
)

var testID = "__test" + fmt.Sprint(rand.Int())
var repo = NewAutostartRepository()

func TestInsertAndFind(t *testing.T) {

}
