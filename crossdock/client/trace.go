// Copyright (c) 2017 Uber Technologies, Inc.
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

package client

import (
	"context"
	"fmt"
	"time"

	"github.com/crossdock/crossdock-go"

	"github.com/jaegertracing/jaeger-client-go/v3/crossdock/common"
	"github.com/jaegertracing/jaeger-client-go/v3/crossdock/log"
	"github.com/jaegertracing/jaeger-client-go/v3/crossdock/thrift/tracetest"
	"github.com/jaegertracing/jaeger-client-go/v3/utils"
)

func (c *Client) trace(t crossdock.T) {
	sampled := str2bool(t.Param(sampledParam))
	baggage := randomBaggage()

	level1 := tracetest.NewStartTraceRequest()
	level1.ServerRole = RoleS1
	level1.Sampled = sampled
	level1.Baggage = baggage
	server1 := t.Param(server1NameParam)

	level2 := tracetest.NewDownstream()
	level2.ServiceName = t.Param(server2NameParam)
	level2.ServerRole = RoleS2
	level2.Host = c.mapServiceToHost(level2.ServiceName)
	level2.Port = c.transport2port(t.Param(server2TransportParam))
	level2.Transport = transport2transport(t.Param(server2TransportParam))
	level1.Downstream = level2

	level3 := tracetest.NewDownstream()
	level3.ServiceName = t.Param(server3NameParam)
	level3.ServerRole = RoleS3
	level3.Host = c.mapServiceToHost(level3.ServiceName)
	level3.Port = c.transport2port(t.Param(server3TransportParam))
	level3.Transport = transport2transport(t.Param(server3TransportParam))
	level2.Downstream = level3

	server1host := c.mapServiceToHost(server1)
	url := fmt.Sprintf("http://%s:%s/start_trace", server1host, c.ServerPortHTTP)
	resp, err := common.PostJSON(context.Background(), url, level1)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	for r := resp; r != nil; r = r.Downstream {
		if r.NotImplementedError != "" {
			t.Skipf(r.NotImplementedError)
			log.Printf("SKIP: %s", r.NotImplementedError)
			return
		}
	}

	traceID := resp.Span.TraceId
	if traceID == "" {
		t.Errorf("Trace ID is empty in S1(%s)", server1)
		return
	}

	success := validateTrace(t, level1.Downstream, resp, server1, 1, traceID, sampled, baggage)
	if success {
		t.Successf("trace checks out")
		log.Printf("PASS")
	}
}

func validateTrace(
	t crossdock.T,
	target *tracetest.Downstream,
	resp *tracetest.TraceResponse,
	service string,
	level int,
	traceID string,
	sampled bool,
	baggage string) bool {

	success := true
	if !equalTraceIDs(traceID, resp.Span.TraceId) {
		t.Errorf("Trace ID mismatch in S%d(%s): expected %s, received %s",
			level, service, traceID, resp.Span.TraceId)
		success = false
	}
	if baggage != resp.Span.Baggage {
		t.Errorf("Baggage mismatch in S%d(%s): expected %s, received %s",
			level, service, baggage, resp.Span.Baggage)
		success = false
	}
	if sampled != resp.Span.Sampled {
		t.Errorf("Sampled mismatch in S%d(%s): expected %t, received %t",
			level, service, sampled, resp.Span.Sampled)
		success = false
	}
	if target != nil {
		if resp.Downstream == nil {
			t.Errorf("Missing downstream in S%d(%s)", level, service)
			success = false
		} else {
			success = validateTrace(t, target.Downstream, resp.Downstream,
				target.Host, level+1, traceID, sampled, baggage) && success
		}
	} else if resp.Downstream != nil {
		t.Errorf("Unexpected downstream in S%d(%s)", level, service)
		success = false
	}
	return success
}

func randomBaggage() string {
	r := utils.NewRand(time.Now().UnixNano())
	n := uint64(r.Int63())
	return fmt.Sprintf("%x", n)
}

func str2bool(v string) bool {
	switch v {
	case "true":
		return true
	case "false":
		return false
	default:
		panic(v + " is not a Boolean")
	}
}

func (c *Client) transport2port(v string) string {
	switch v {
	case transportHTTP:
		return c.ServerPortHTTP
	case transportDummy:
		return "9999"
	default:
		panic("Unknown protocol " + v)
	}
}

func transport2transport(v string) tracetest.Transport {
	switch v {
	case transportHTTP:
		return tracetest.Transport_HTTP
	case transportDummy:
		return tracetest.Transport_DUMMY
	default:
		panic("Unknown protocol " + v)
	}
}

func equalTraceIDs(id1, id2 string) bool {
	return padTraceID(id1) == padTraceID(id2)
}

func padTraceID(id string) string {
	if len(id) == 16 {
		return id
	}
	id = "0000000000000000" + id
	return id[len(id)-16:]
}
