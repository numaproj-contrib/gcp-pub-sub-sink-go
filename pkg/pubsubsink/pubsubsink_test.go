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
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/pubsub"
	sinksdk "github.com/numaproj/numaflow-go/pkg/sinker"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	"github.com/shubhamdixit863/gcp-pub-sub-sink-go/pkg/mocks"
)

const TopicID = "pubsub-test"
const Project = "pubsub-local-test"
const PUB_SUB_EMULATOR_HOST = "localhost:8681"

var (
	pubsubClient *pubsub.Client
	resource     *dockertest.Resource
	pool         *dockertest.Pool
)

// getTopic checks if the specified topic name exists and returns it .
func getTopic(ctx context.Context, client *pubsub.Client, topicID string) (*pubsub.Topic, error) {
	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		if topic, err = client.CreateTopic(ctx, topicID); err != nil {
			return nil, err
		}
	}
	return topic, nil
}

func TestMain(m *testing.M) {
	var err error
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker ;is it running ? %s", err)
	}
	pool = p
	// Check if pubsub container is already running
	containers, err := pool.Client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		log.Fatalf("could not list containers %s", err)
	}
	pubSubRunning := false
	for _, container := range containers {
		fmt.Println(container)
		for _, name := range container.Names {
			if strings.Contains(name, "google-cloud-cli") {
				pubSubRunning = true
				break
			}
		}
		if pubSubRunning {
			break
		}
	}
	if !pubSubRunning {
		// Start pubsub container if not already running
		opts := dockertest.RunOptions{
			Repository:   "thekevjames/gcloud-pubsub-emulator",
			Tag:          "latest",
			ExposedPorts: []string{"8681"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"8681": {
					{HostIP: "127.0.0.1", HostPort: "8681"},
				},
			},
		}
		resource, err = pool.RunWithOptions(&opts)
		if err != nil {
			_ = pool.Purge(resource)
			log.Fatalf("could not start resource %s", err)
		}
	}
	err = os.Setenv("PUBSUB_EMULATOR_HOST", PUB_SUB_EMULATOR_HOST)
	if err != nil {
		log.Fatalf("error -%s", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := pool.Retry(func() error {
		pubsubClient, err = pubsub.NewClient(ctx, Project)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
		return nil
	}); err != nil {
		if resource != nil {
			_ = pool.Purge(resource)
		}
		log.Fatalf("could not connect to gcloud pubsub %s", err)
	}
	defer pubsubClient.Close()
	defer cancel()
	code := m.Run()
	if resource != nil {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Couln't purge resource %s", err)
		}
	}
	os.Exit(code)
}

func TestPubSubSink_Sink(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	topic, err := getTopic(ctx, pubsubClient, TopicID)
	assert.Nil(t, err)
	pubSubSink := NewPubSubSink(topic)
	ch := make(chan sinksdk.Datum, 20)
	closeCh := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			ch <- mocks.Payload{
				Data: "pubsub test",
			}
		}
		close(ch)
		close(closeCh)
	}()
	<-closeCh
	response := pubSubSink.Sink(ctx, ch)
	assert.Equal(t, 10, len(response.Items()))
}
