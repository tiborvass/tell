package functions

import (
	"fmt"
	"reflect"

	"github.com/docker/libchan"
)

type Function struct {
	Name   string
	Args   []interface{}
	Return libchan.Sender
}

func Export(s interface{}, receiver libchan.Receiver, ch chan error) error {
	if receiver == nil {
		return fmt.Errorf("receiver cannot be nil")
	}

	// Access struct
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("ExportFunctions only accepts structs")
	}

	// In case something goes wrong, panic if no error channel was provided.
	// Otherwise, send error to that channel.
	fatal := func(err error) {
		if ch == nil {
			panic(err)
		}
		ch <- err
	}

	// Start listening on receiver
	go func() {
		for {
			// Receive the function message
			var msg Function
			if err := receiver.Receive(&msg); err != nil {
				fatal(fmt.Errorf("server: receive err: %v", err))
				continue
			}

			// Look up the function name
			// The name has to match a method of the struct passed to ExportFunctions
			method := v.MethodByName(msg.Name)
			if !method.IsValid() {
				fatal(fmt.Errorf("server: could not find function %q", msg.Name))
				continue
			}

			// Convert input arguments from the message to the function arguments
			var in []reflect.Value
			in = make([]reflect.Value, len(msg.Args))
			for i, arg := range msg.Args {
				in[i] = reflect.ValueOf(arg).Convert(method.Type().In(i))
			}

			results := method.Call(in)

			// Build the return values from the results of the call
			out := make([]interface{}, len(results))
			for i, result := range results {
				out[i] = result.Interface()
			}

			// Send the return values to the return channel specified in the function message
			if err := msg.Return.Send(&out); err != nil {
				fatal(fmt.Errorf("server: send err: %v", err))
				continue
			}
			if err := msg.Return.Close(); err != nil {
				fatal(fmt.Errorf("server: close err: %v", err))
				continue
			}
		}
	}()
	return nil
}

func PairFromStruct(s interface{}, sender libchan.Sender) error {
	if sender == nil {
		return fmt.Errorf("sender cannot be nil")
	}

	// Access struct
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expecting struct as first argument")
	}
	t := v.Type()

	// Iterate through each struct field, that should all be nil functions
	for i := 0; i < t.NumField(); i++ {
		// Associate each function to a call to a remote function with the same name
		if err := Pair(t.Field(i).Name, v.Field(i).Addr(), sender); err != nil {
			return err
		}
	}
	return nil
}

func Pair(name string, fn interface{}, sender libchan.Sender) error {
	if sender == nil {
		return fmt.Errorf("sender cannot be nil")
	}

	// Access function
	v, ok := fn.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(fn)
	}
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("fn needs to be a pointer to a function")
	}
	v = v.Elem()
	if v.Kind() != reflect.Func {
		return fmt.Errorf("fn needs to be a pointer to a function")
	}
	if !v.CanSet() {
		return fmt.Errorf("fn needs to be settable")
	}
	t := v.Type()

	// Create a callable function of the same type, on the fly
	v.Set(reflect.MakeFunc(t, func(in []reflect.Value) []reflect.Value {
		n := t.NumOut()

		// create the pipe for return values
		retR, retW := libchan.Pipe()

		// construct arguments to be put in the function message
		args := make([]interface{}, len(in))
		for i, e := range in {
			args[i] = e.Interface()
		}

		// send function message
		if err := sender.Send(&Function{
			Name:   name,
			Args:   args,
			Return: retW,
		}); err != nil {
			panic(fmt.Errorf("%s: error sending: %v", name, err))
		}

		// wait for response message
		var msg []interface{}
		if err := retR.Receive(&msg); err != nil {
			panic(fmt.Errorf("%s: error receiving: %v", name, err))
		}

		if len(msg) != n {
			panic(fmt.Errorf("%s: number of return values differ is %d, but should be %d", name, len(msg), n))
		}

		// convert return values from message to the function's return types
		results := make([]reflect.Value, n)
		for i := 0; i < n; i++ {
			x := reflect.ValueOf(msg[i])
			newT := t.Out(i)
			if !x.Type().ConvertibleTo(newT) {
				panic(fmt.Errorf("%s: could not convert return value #%d from %s to %s", name, i, x.Type(), newT))
			}
			results[i] = x.Convert(t.Out(i))
		}

		return results
	}))
	return nil
}
