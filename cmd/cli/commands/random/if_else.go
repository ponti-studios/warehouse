package random

import "fmt"

func if_else() {
	x := 5

	if x == 5 {
		fmt.Println("x is 5")
	}

	if x != 5 {
		fmt.Println("x is not 5")
	} else {
		fmt.Println("x is 5")
	}
}
