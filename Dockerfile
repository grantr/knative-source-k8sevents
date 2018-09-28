# Copyright 2018 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-1.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang AS builder

# copy the source
ADD . /go/src/github.com/knative/source-k8sevents
WORKDIR /go/src/github.com/knative/source-k8sevents

# build the sample
RUN CGO_ENABLED=0 go build -o /go/bin/receive-adapter .

FROM golang:alpine

EXPOSE 8080
COPY --from=builder /go/bin/receive-adapter /app/receive-adapter

ENTRYPOINT ["/app/receive-adapter"]
