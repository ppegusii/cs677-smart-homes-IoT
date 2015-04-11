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
