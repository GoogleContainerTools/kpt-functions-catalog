# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
IMG ?=gcr.io/sunilarora-sandbox/httpbin-gen:latest

fn:
	go build -o bin/httpbin-gen ./

run-fn:
	$(HOME)/go/bin/kpt fn run examples/pkg --enable-exec --exec-path bin/httpbin-gen

clean:
	rm -f bin/httpbin-gen

docker:
	docker build -t ${IMG} -f Dockerfile .

docker-push:
	docker push ${IMG}