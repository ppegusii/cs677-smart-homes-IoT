// This file defines the basic selection menu for clock sync, logical clocks or no ordering
// Based on the value selected the corresponding middleware stub is started

package ordermw

import (
	"github.com/ppegusii/cs677-smart-homes-IoT/api"
	"log"
)

func GetOrderingMiddleware(o api.Ordering, id int, ip string, port string) api.OrderingMiddlewareInterface {
	switch o {
	case api.ClockSync:
		return BClockNewDummy(id, ip, port)
	case api.LogicalClock:
		return NewLogical(id, ip, port)
	case api.NoOrder:
		return NewDummy(id, ip, port)
	default:
		log.Printf("Invalid ordering: %d", o)
		return nil
	}
}
