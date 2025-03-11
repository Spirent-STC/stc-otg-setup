package gosnappi_examples

import (
	"fmt"
	"os"
	"strings"
)

var (
	PORT1     = "//10.61.37.185/1/1"
	PORT2     = "//10.61.37.185/1/2"
	OTGSERVER = "localhost:50051"
)

func init() {
	if p1 := os.Getenv("PORT1"); p1 != "" {
		PORT1 = "//" + strings.Trim(p1, "/")
	}
	if p2 := os.Getenv("PORT2"); p2 != "" {
		PORT2 = "//" + strings.Trim(p2, "/")
	}

	if otgserver := os.Getenv("OTGSERVER"); otgserver != "" {
		OTGSERVER = otgserver
	}

	fmt.Printf("PORT1 = %v, PORT2 = %v, OTGSERVER = %v", PORT1, PORT2, OTGSERVER)
	fmt.Println()
}
