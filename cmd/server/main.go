package server

import (
	"fmt"
	"time"
)

func AboutMe() string {
	return fmt.Sprintf(`Copyright© 2021 - %d Nmid-Registry(http://www.niansong.top). Nmid-Registry is Nmid Register Center All rights reserved.Apache License 2.0.`, time.Now().Year())
}
