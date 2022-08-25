// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-labels/transformer"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func run() error {
	// Register js function processResourceList to the globals.
	js.Global().Set("processResourceList", resourceListProcessorWrapper())
	// We need to ensure that the Go program is running when JavaScript calls it.
	// Otherwise, it will complain the Go program has already exited.
	<-make(chan bool)
	return nil
}

func transformLabels(input []byte) ([]byte, error) {
	return fn.Run(fn.ResourceListProcessorFunc(transformer.SetLabels), []byte(input))
}

func resourceListProcessorWrapper() js.Func {
	// TODO: figure out a better way to surface a golang error to JS environment.
	// Currently error is surfaced as a string.
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return "Invalid number of arguments passed"
		}
		input := args[0].String()
		transformed, err := transformLabels([]byte(input))
		if err != nil {
			return fmt.Errorf("unable to process resource list:", err.Error())
		}
		return string(transformed)
	})
	return jsonFunc
}
