/**
 * Copyright 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scbe

import (
	"github.com/IBM/ubiquity/resources"
)

type ScbeRestClientGen func(resources.ConnectionInfo) (ScbeRestClient, error)

var globalScbeRestClientGen ScbeRestClientGen = nil

func InitScbeRestClientGen(gen ScbeRestClientGen) func() {
	if globalScbeRestClientGen != nil {
		panic("globalScbeRestClientGen already initialized")
	}
	globalScbeRestClientGen = gen
	return func() { globalScbeRestClientGen = nil }
}

func newScbeRestClientGen(conInfo resources.ConnectionInfo) (ScbeRestClient, error) {
	if globalScbeRestClientGen != nil {
		return globalScbeRestClientGen(conInfo)
	}
	return NewScbeRestClient(conInfo)
}
