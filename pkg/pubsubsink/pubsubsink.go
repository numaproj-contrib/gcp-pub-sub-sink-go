/*
Copyright 2022 The Numaproj Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package pubsubsink

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
	sinksdk "github.com/numaproj/numaflow-go/pkg/sinker"
)

type PubSubSink struct {
	topic *pubsub.Topic
}

func NewPubSubSink(topic *pubsub.Topic) *PubSubSink {
	return &PubSubSink{topic: topic}
}

func (p *PubSubSink) Sink(ctx context.Context, datumStreamCh <-chan sinksdk.Datum) sinksdk.Responses {
	responses := sinksdk.ResponsesBuilder()
	var results []*pubsub.PublishResult
	for datum := range datumStreamCh {
		result := p.topic.Publish(ctx, &pubsub.Message{
			ID:   datum.ID(),
			Data: datum.Value(),
		})
		results = append(results, result)
	}
	for _, res := range results {
		id, err := res.Get(ctx)
		if err != nil {
			responses = append(responses, sinksdk.Response{
				ID:      id,
				Success: false,
				Err:     err.Error(),
			})
			log.Printf("error publishing messages- %s", err)
			continue
		}
		responses = append(responses, sinksdk.Response{
			ID:      id,
			Success: true,
			Err:     "",
		})
	}
	return responses
}
