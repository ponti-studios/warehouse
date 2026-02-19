// types_and_operators.go
// This file provides an overview of types and operators in Go, complete with comments and examples for learning.

package tools

import (
	"fmt"
)

// In Go, every variable has a type, which defines the kind of data it can hold and the operations
// that can be performed on it. Go is a statically typed language, meaning that types are
// checked at compile time.

func types_and_operators() {
	// Basic Types in Go:
	// 1. Boolean Type
	// 2. Numeric Types (Integers, Floating-point numbers, Complex numbers)
	// 3. String Type

	// 1. Boolean Type:
	// A boolean type is used to represent truth values. It can hold one of two values: true or false.
	var isGoFun bool = true
	fmt.Println("isGoFun:", isGoFun)

	// Boolean operators:
	// && (AND), || (OR), ! (NOT)
	a := true
	b := false
	fmt.Println("a && b:", a && b) // false, because one operand is false
	fmt.Println("a || b:", a || b) // true, because at least one operand is true
	fmt.Println("!a:", !a)         // false, because the NOT operator inverts the value

	// 2. Numeric Types:
	// Go has several numeric types, including integers, floating-point numbers, and complex numbers.

	// Integer Types:
	// As previously mentioned, Go has both signed and unsigned integer types with varying sizes.
	var x int = 42       // Default integer type, size depends on the machine architecture
	var y int8 = -128    // 8-bit signed integer (-128 to 127)
	var z uint16 = 65535 // 16-bit unsigned integer (0 to 65535)

	fmt.Println("x:", x)
	fmt.Println("y:", y)
	fmt.Println("z:", z)

	// Integer Operators:
	// Arithmetic: +, -, *, /, %
	// Comparison: ==, !=, <, >, <=, >=
	fmt.Println("x + 10:", x+10) // Addition
	fmt.Println("y - 1:", y-1)   // Subtraction
	fmt.Println("z * 2:", z*2)   // Multiplication
	fmt.Println("x / 2:", x/2)   // Division
	fmt.Println("x % 5:", x%5)   // Modulus (remainder of division)

	// Bitwise Operators (only applicable to integer types):
	// & (AND), | (OR), ^ (XOR), &^ (AND NOT), << (Left Shift), >> (Right Shift)
	fmt.Println("x & 1:", x&1)   // Bitwise AND
	fmt.Println("x | 1:", x|1)   // Bitwise OR
	fmt.Println("x ^ 1:", x^1)   // Bitwise XOR
	fmt.Println("x << 1:", x<<1) // Left shift (equivalent to multiplying by 2)
	fmt.Println("x >> 1:", x>>1) // Right shift (equivalent to dividing by 2)

	// Floating-Point Types:
	// float32 and float64 as previously discussed. These types represent numbers with fractional components.
	var f32 float32 = 3.14
	var f64 float64 = 2.718281828459045

	fmt.Println("f32:", f32)
	fmt.Println("f64:", f64)

	// Arithmetic operations are similar to integers:
	fmt.Println("f32 + f64:", float64(f32)+f64) // Note: type conversion is required

	// Complex Types:
	// Go also has built-in support for complex numbers, which have both real and imaginary components.
	var c complex64 = 1 + 2i  // Complex number with float32 real and imaginary parts
	var d complex128 = 3 + 4i // Complex number with float64 real and imaginary parts

	fmt.Println("c:", c)
	fmt.Println("d:", d)

	// Complex number operations:
	// + (Addition), - (Subtraction), * (Multiplication), / (Division)
	fmt.Println("c + d:", complex128(c)+d) // Type conversion required for operations between different complex types
	fmt.Println("Real part of c:", real(c))
	fmt.Println("Imaginary part of c:", imag(c))

	// 3. String Type:
	// A string is a sequence of bytes (typically representing UTF-8 encoded text).
	// Strings are immutable in Go, meaning that once created, they cannot be changed.

	var greeting string = "Hello, Go!"
	fmt.Println("greeting:", greeting)

	// String operations:
	// + (Concatenation), len() (Length), indexing and slicing
	name := "World"
	fullGreeting := greeting + " " + name
	fmt.Println("fullGreeting:", fullGreeting)                // Concatenation
	fmt.Println("Length of fullGreeting:", len(fullGreeting)) // Length
	fmt.Println("First character:", fullGreeting[0])          // Indexing (returns byte)
	fmt.Println("Substring:", fullGreeting[7:])               // Slicing (returns substring)

	// Note: Strings can be converted to byte slices for more advanced manipulations.
	byteSlice := []byte(fullGreeting)
	fmt.Println("Byte slice:", byteSlice)
}
