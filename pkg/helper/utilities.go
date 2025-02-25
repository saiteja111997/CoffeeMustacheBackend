package helper

import (
	"coffeeMustacheBackend/pkg/structures"
	"os"
)

func IsLambda() bool {
	if lambdaTaskRoot := os.Getenv("LAMBDA_TASK_ROOT"); lambdaTaskRoot != "" {
		return true
	}
	return false
}

func IsValid(addedvia structures.CartInsertType) bool {
	switch structures.CartInsertType(addedvia) {
	case structures.Direct, structures.FromCuratedCart, structures.CrossSellFocus, structures.TopPicks, structures.UpgradeCartAi, structures.CrossSellCheckout:
		return true
	default:
		return false
	}
}
