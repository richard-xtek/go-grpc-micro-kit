// Copyright (c) 2019 The Jaeger Authors.
// Copyright (c) 2017-2018 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package recoveryhandler

import (
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/sirupsen/logrus"
)

// recoveryWrapper wraps a logger into a gorilla RecoveryLogger
type recoveryWrapper struct {
	logger *logrus.Entry
}

// Println logs an error message with the given fields
func (z recoveryWrapper) Println(args ...interface{}) {
	z.logger.Error(fmt.Sprint(args...))
}

// NewRecoveryHandler returns an http.Handler that recovers on panics
func NewRecoveryHandler(logger *logrus.Entry, printStack bool) func(h http.Handler) http.Handler {
	zWrapper := recoveryWrapper{logger}
	return handlers.RecoveryHandler(handlers.RecoveryLogger(zWrapper), handlers.PrintRecoveryStack(printStack))
}
