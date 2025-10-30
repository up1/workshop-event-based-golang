package s2

import "s1"

func Greet() string {
	return s1.Hello() + " and s2"
}
