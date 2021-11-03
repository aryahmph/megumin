package service

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type GameService interface {
	RandomGuestNumber()
	//GuessNumber(number int, timestamp uint64) bool
	Valid(timestamp uint64) bool
	GetNumber() int
	End(timestamp uint64)
}

type GameServiceImpl struct {
	RWMutex   *sync.RWMutex
	number    int
	timestamp uint64
}

func NewGameServiceImpl(RWMutex *sync.RWMutex) *GameServiceImpl {
	return &GameServiceImpl{RWMutex: RWMutex, number: -1, timestamp: 0}
}

func (service *GameServiceImpl) RandomGuestNumber() {
	service.RWMutex.Lock()
	service.number = 1 + rand.Intn(10)
	service.timestamp = uint64(time.Now().Add(20 * time.Second).Unix())
	service.RWMutex.Unlock()
}

func (service *GameServiceImpl) Valid(timestamp uint64) bool {
	service.RWMutex.RLock()
	var isValid bool
	isValid = service.number != -1
	isValid = service.timestamp != 0 && timestamp <= service.timestamp
	fmt.Println(service.number)
	fmt.Println(service.timestamp)
	service.RWMutex.RUnlock()
	return isValid
}

func (service *GameServiceImpl) End(timestamp uint64) {
	service.number = -1
	service.timestamp = timestamp
}

func (service *GameServiceImpl) GetNumber() int {
	service.RWMutex.RLock()
	number := service.number
	service.RWMutex.RUnlock()
	return number
}
