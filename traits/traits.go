package traits

import "reflect"

type TraitHolder interface {
	TraitData() *TraitData
}

type TraitData struct {
	traits map[reflect.Type]any
	// map[event type]map[trait type]function
	listeners map[reflect.Type]map[reflect.Type]any
	events    map[reflect.Type]bool
}

func NewData(supportedEvents ...reflect.Type) *TraitData {
	events := make(map[reflect.Type]bool)
	for _, event := range supportedEvents {
		events[event] = true
	}

	return &TraitData{
		traits:    make(map[reflect.Type]any),
		listeners: make(map[reflect.Type]map[reflect.Type]any),
		events:    events,
	}
}

func Set(target TraitHolder, trait any) {
	val := reflect.ValueOf(trait)
	ty := val.Type()
	traitData := target.TraitData()

	for i := 0; i < val.NumMethod(); i++ {
		method := val.Method(i)
		methodTy := method.Type()
		if methodTy.NumIn() != 1 {
			continue
		}

		paramTy := methodTy.In(0)
		if _, ok := traitData.events[paramTy]; !ok {
			continue
		}

		listeners, ok := traitData.listeners[paramTy]
		if !ok {
			listeners = make(map[reflect.Type]any)
			traitData.listeners[paramTy] = listeners
		}

		listeners[ty] = method.Interface()
	}

	traitData.traits[ty] = trait
}

func Unset(target TraitHolder, trait any) {
	val := reflect.ValueOf(trait)
	traitData := target.TraitData()

	for i := 0; i < val.NumMethod(); i++ {
		method := val.Method(i)
		methodTy := method.Type()
		if methodTy.NumIn() != 1 {
			continue
		}

		paramTy := methodTy.In(0)
		listeners, ok := traitData.listeners[paramTy]
		if !ok {
			continue
		}

		delete(listeners, val.Type())
	}
}

func Get[T any](target TraitHolder) *T {
	trait, ok := target.TraitData().traits[reflect.TypeOf((*T)(nil))]
	if !ok {
		return nil
	} else {
		return trait.(*T)
	}
}

func CallEvent[T any](traitData *TraitData, event *T) {
	listeners, ok := traitData.listeners[reflect.TypeOf(event)]
	if !ok {
		return
	}

	for _, fn := range listeners {
		fn.(func(*T))(event)
	}
}
