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

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
	sinksdk "github.com/numaproj/numaflow-go/pkg/sinker"

	"github.com/numaproj-contrib/gcp-pub-sub-sink-go/pkg/pubsubsink"
)

// getTopic checks if the specified topic name exists and returns it.
func getTopic(ctx context.Context, client *pubsub.Client, topicID string) (*pubsub.Topic, error) {
	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("topic does not exist: %s", topicID)
	}
	return topic, nil
}

func main() {
	topicId := os.Getenv("TOPIC_ID")
	projectId := os.Getenv("PROJECT_ID")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		log.Fatalf("error in creating pubsub client: %s", err)
	}
	defer client.Close()
	topic, err := getTopic(ctx, client, topicId)
	if err != nil {
		log.Fatalf("error in getting topic: %s", err)
	}
	pubSubSink := pubsubsink.NewPubSubSink(topic)
	if err := sinksdk.NewServer(pubSubSink).Start(ctx); err != nil {
		log.Fatalf("failed to start sinker server: %s", err)
	}
}
