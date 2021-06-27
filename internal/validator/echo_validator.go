package validator

import (
	"reflect"
	"sync"

	"github.com/go-playground/validator/v10"
)

type EchoValidator struct {
	once     sync.Once
	validate *validator.Validate
}

func (v *EchoValidator) Validate(obj interface{}) error {
	if kindOfData(obj) == reflect.Struct {
		v.lazyInit()

		if err := v.validate.Struct(obj); err != nil {
			return err
		}
	}

	return nil
}

func (v *EchoValidator) Engine() *EchoValidator {
	v.lazyInit()
	return v
}

func (v *EchoValidator) lazyInit() {
	v.once.Do(func() {
		v.validate = validator.New()
		// add any custom validations etc. here
	})
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}