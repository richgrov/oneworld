package traits_test

import (
	"reflect"
	"testing"

	"github.com/richgrov/oneworld/traits"
)

type TestEvent struct {
	message string
}

type UnregisteredEvent struct {
	message string
}

type TestTrait struct {
	field                string
	eventRecieved        bool
	invalidEventRecieved bool
}

func (t *TestTrait) OnEvent(e *TestEvent) {
	if e.message == "secret" {
		t.eventRecieved = true
	}
}

func (t *TestTrait) OnInvalidEvent(e *UnregisteredEvent) {
	t.invalidEventRecieved = true
}

type TestHolder struct {
	traitData *traits.TraitData
}

func (h *TestHolder) TraitData() *traits.TraitData {
	return h.traitData
}

func TestSet(t *testing.T) {
	trait := &TestTrait{
		field:                "hello",
		eventRecieved:        false,
		invalidEventRecieved: false,
	}
	holder := &TestHolder{
		traitData: traits.NewData(reflect.TypeOf(&TestEvent{})),
	}

	// Trait should not be added yet
	traitInstance := traits.Get[TestTrait](holder)
	if traitInstance != nil {
		t.Errorf("got trait instance before setting trait: %+v", traitInstance)
	}

	// Make sure trait is retrievable
	traits.Set(holder, trait)
	traitInstance = traits.Get[TestTrait](holder)
	if traitInstance == nil {
		t.Errorf("got nil trait instance after setting trait")
	}

	// Ensure trait fields are unchanged
	if traitInstance.field != "hello" || traitInstance.eventRecieved || traitInstance.invalidEventRecieved {
		t.Errorf("trait has incorrect fields: %+v", traitInstance)
	}

	// Call an event and make sure the trait recieved it
	event := &TestEvent{
		message: "secret",
	}

	traits.CallEvent(holder.traitData, event)
	if !traitInstance.eventRecieved {
		t.Error("trait should've recieved event")
	}

	// Ensure that calling unregistered events doesn't get passed to the trait
	unregisteredEvent := &UnregisteredEvent{}
	traits.CallEvent(holder.traitData, unregisteredEvent)
	if traitInstance.invalidEventRecieved {
		t.Error("trait recieved an event that wasn't registered")
	}
}
