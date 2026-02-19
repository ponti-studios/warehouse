// comparisons.go
// This file provides a comprehensive overview of comparisons and numeric types in Go.
// It includes detailed comments and examples for learning purposes.

package tools

import (
	"fmt"
)

// Go supports comparison operations between variables, but only if they are of the same type.
// This is important to ensure type safety and avoid unexpected results during comparison.

// Comparison Operators:
// == : Equal to
// != : Not equal to
// <  : Less than
// >  : Greater than
// <= : Less than or equal to
// >= : Greater than or equal to

func comparisons() {
	// Example: Comparing integers of the same type
	a := 10
	b := 20
	fmt.Println("a == b:", a == b) // false, because 10 is not equal to 20
	fmt.Println("a < b:", a < b)   // true, because 10 is less than 20

	// Trying to compare variables of different types would result in a compilation error
	// Uncommenting the following line would cause a compilation error:
	// fmt.Println("a == 10.0:", a == 10.0) // Error: cannot compare int and float

	// Numeric Types in Go:
	// Go provides both architecture-dependent and architecture-independent numeric types.

	// Architecture-Dependent Types:
	// These types have sizes that depend on the machine architecture (32-bit or 64-bit).

	// int: This is the default integer type. Its size is 32 bits on a 32-bit machine
	// and 64 bits on a 64-bit machine.
	var x int = 42
	fmt.Println("x:", x)

	// uint: This is the unsigned version of int, which means it can only hold non-negative values.
	var y uint = 42
	fmt.Println("y:", y)

	// uintptr: This is an unsigned integer type large enough to store the uninterpreted bits of a pointer value.
	// It's used internally by the runtime and is not typically used in regular code.
	var z uintptr = 0xDEADBEEF
	fmt.Println("z:", z)

	// Architecture-Independent Types:
	// These types have fixed sizes, which are indicated by their names.

	// Integers:
	var i8 int8 = 127                   // 8-bit signed integer (-128 to 127)
	var i16 int16 = 32767               // 16-bit signed integer (-32768 to 32767)
	var i32 int32 = 2147483647          // 32-bit signed integer (-2^31 to 2^31-1)
	var i64 int64 = 9223372036854775807 // 64-bit signed integer (-2^63 to 2^63-1)

	fmt.Println("i8:", i8)
	fmt.Println("i16:", i16)
	fmt.Println("i32:", i32)
	fmt.Println("i64:", i64)

	// Unsigned Integers:
	var ui8 uint8 = 255                    // 8-bit unsigned integer (0 to 255)
	var ui16 uint16 = 65535                // 16-bit unsigned integer (0 to 65535)
	var ui32 uint32 = 4294967295           // 32-bit unsigned integer (0 to 2^32-1)
	var ui64 uint64 = 18446744073709551615 // 64-bit unsigned integer (0 to 2^64-1)

	fmt.Println("ui8:", ui8)
	fmt.Println("ui16:", ui16)
	fmt.Println("ui32:", ui32)
	fmt.Println("ui64:", ui64)

	// Floating-Point Numbers:
	// Go has two floating-point types: float32 and float64, which are based on the IEEE-754 standard.

	// float32: Single-precision floating-point number
	// It is accurate to about 7 decimal places.
	var f32 float32 = 3.14159
	fmt.Println("f32:", f32)

	// float64: Double-precision floating-point number
	// It is accurate to about 15 decimal places and should be used whenever possible.
	// The "math" package functions expect float64 types.
	var f64 float64 = 3.141592653589793
	fmt.Println("f64:", f64)

	// NOTE: Go does not have a `float` type. Instead, you should use `float32` or `float64` explicitly.
}
