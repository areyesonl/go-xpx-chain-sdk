package subscribers

import (
	"github.com/proximax-storage/go-xpx-chain-sdk/sdk"
)

type (
	UnconfirmedAddedHandler func(sdk.Transaction) bool
	UnconfirmedAdded        interface {
		AddHandlers(address *sdk.Address, handlers ...*UnconfirmedAddedHandler) error
		RemoveHandlers(address *sdk.Address, handlers ...*UnconfirmedAddedHandler) (bool, error)
		HasHandlers(address *sdk.Address) bool
		GetHandlers(address *sdk.Address) []*UnconfirmedAddedHandler
		GetAddresses() []string
	}

	unconfirmedAddedImpl struct {
		newSubscriberCh    chan *unconfirmedAddedSubscription
		removeSubscriberCh chan *unconfirmedAddedSubscription
		subscribers        map[string][]*UnconfirmedAddedHandler
	}
	unconfirmedAddedSubscription struct {
		address  *sdk.Address
		handlers []*UnconfirmedAddedHandler
		resultCh chan bool
	}
)

func NewUnconfirmedAdded() UnconfirmedAdded {

	p := &unconfirmedAddedImpl{
		subscribers:        make(map[string][]*UnconfirmedAddedHandler),
		newSubscriberCh:    make(chan *unconfirmedAddedSubscription),
		removeSubscriberCh: make(chan *unconfirmedAddedSubscription),
	}
	go p.handleNewSubscription()
	return p
}

func (e *unconfirmedAddedImpl) handleNewSubscription() {
	for {
		select {
		case s := <-e.newSubscriberCh:
			e.addSubscription(s)
		case s := <-e.removeSubscriberCh:
			e.removeSubscription(s)
		}
	}
}

func (e *unconfirmedAddedImpl) addSubscription(s *unconfirmedAddedSubscription) {

	if _, ok := e.subscribers[s.address.Address]; !ok {
		e.subscribers[s.address.Address] = make([]*UnconfirmedAddedHandler, 0)
	}
	for i := 0; i < len(s.handlers); i++ {
		e.subscribers[s.address.Address] = append(e.subscribers[s.address.Address], s.handlers[i])
	}
}

func (e *unconfirmedAddedImpl) removeSubscription(s *unconfirmedAddedSubscription) {

	if external, ok := e.subscribers[s.address.Address]; !ok || len(external) == 0 {
		s.resultCh <- false
	}

	itemCount := len(e.subscribers[s.address.Address])
	for _, removeHandler := range s.handlers {
		for index, currentHandlers := range e.subscribers[s.address.Address] {
			if removeHandler == currentHandlers {
				e.subscribers[s.address.Address] = append(e.subscribers[s.address.Address][:index],
					e.subscribers[s.address.Address][index+1:]...)
			}
		}
	}

	s.resultCh <- itemCount != len(e.subscribers[s.address.Address])
}

func (e *unconfirmedAddedImpl) AddHandlers(address *sdk.Address, handlers ...*UnconfirmedAddedHandler) error {

	if len(handlers) == 0 {
		return nil
	}

	e.newSubscriberCh <- &unconfirmedAddedSubscription{
		address:  address,
		handlers: handlers,
	}
	return nil
}

func (e *unconfirmedAddedImpl) RemoveHandlers(address *sdk.Address, handlers ...*UnconfirmedAddedHandler) (bool, error) {
	if len(handlers) == 0 {
		return false, nil
	}

	resCh := make(chan bool)
	e.removeSubscriberCh <- &unconfirmedAddedSubscription{
		address:  address,
		handlers: handlers,
		resultCh: resCh,
	}

	return <-resCh, nil
}

func (e *unconfirmedAddedImpl) HasHandlers(address *sdk.Address) bool {
	return len(e.subscribers[address.Address]) > 0 && e.subscribers[address.Address] != nil
}

func (e *unconfirmedAddedImpl) GetHandlers(address *sdk.Address) []*UnconfirmedAddedHandler {

	if res, ok := e.subscribers[address.Address]; ok && res != nil {
		return res
	}

	return nil
}

func (e *unconfirmedAddedImpl) GetAddresses() []string {
	addresses := make([]string, 0, len(e.subscribers))
	for addr := range e.subscribers {
		addresses = append(addresses, addr)
	}

	return addresses
}
