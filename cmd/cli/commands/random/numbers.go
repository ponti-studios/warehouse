package random

import (
	"fmt"
)

func numbers() {
	const int_64 int64 = 127
	const int_64_2 int64 = 12789
	const float_64 float64 = 10.12365
	fmt.Println(int_64)
	fmt.Println(float_64)
	fmt.Println(int_64 + int_64_2)
	// fmt.Println(int_64 + float_64) // invalid operation
}
