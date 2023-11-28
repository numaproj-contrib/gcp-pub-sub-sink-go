# GCP Pub/Sub Sink for Numaflow

The GCP Pub/Sub Sink is a custom user-defined sink for [Numaflow](https://numaflow.numaproj.io/) that enables the integration of Google Cloud Pub/Sub as a sink within your Numaflow pipelines. This integration facilitates the seamless transfer of data from Numaflow pipelines into Google Cloud Pub/Sub topics.

## Quick Start
This quick start guide will help you in setting up a GCP Pub/Sub sink in a Numaflow pipeline.

### Prerequisites
* [Install Numaflow on your Kubernetes cluster](https://numaflow.numaproj.io/quick-start/)
* Access to a Google Cloud Platform (GCP) account with Pub/Sub enabled.
* [GCP CLI configured with access to your GCP project](https://cloud.google.com/sdk/docs/install)

### Step-by-step Guide

#### 1. Set Up Google Cloud Pub/Sub

Using GCP Console or the GCP CLI, create a new Pub/Sub topic and subscription in your GCP project.

#### 2. Deploy a Numaflow Pipeline with GCP Pub/Sub Sink

- Save the following Kubernetes manifest to a file (e.g., `pubsub-sink-pipeline.yaml`)
- Modify the `PROJECT_ID`, `TOPIC_ID`, and `SUBSCRIPTION_ID` as per your GCP Pub/Sub setup
- For production, leave `PUBSUB_EMULATOR_HOST` as a blank string. Use it only for testing with an emulator.


```yaml
apiVersion: numaflow.numaproj.io/v1alpha1
kind: Pipeline
metadata:
  name: gcp-pubsub-sink
spec:
  vertices:
    - name: in
      source:
        generator:
          rpu: 10
          duration: 1s
          msgSize: 100
    - name: out
      sink:
        udsink:
          container:
            image: "quay.io/numaio/numaflow-go/gcloud-pubsub-sink:latest"
            env:
              - name: PROJECT_ID
                value: "your-gcp-project-id"
              - name: TOPIC_ID
                value: "your-pubsub-topic-id"
              - name: SUBSCRIPTION_ID
                value: "your-pubsub-subscription-id"
              - name: PUBSUB_EMULATOR_HOST   # For production keep it blank string
                value: ""
  edges:
    - from: in
      to: out
```

Replace `your-gcp-project-id`, `your-pubsub-topic-id`, and `your-pubsub-subscription-id` with your actual GCP project ID, Pub/Sub topic ID, and subscription ID respectively.Keep PUBSUB_EMULATOR_HOST as blank so it can connect with gcp host in cloud
for local tests you can pass the host ip for emulator.

Then apply it to your cluster:
```bash
kubectl apply -f pubsub-sink-pipeline.yaml
```

#### 3. Verify the Pub/Sub Sink

Verify that messages are being published to the specified Pub/Sub topic through the GCP Console or the GCP CLI.

#### 4. Clean up

To delete the Numaflow pipeline:
```bash
kubectl delete -f pubsub-sink-pipeline.yaml
```

To delete the Pub/Sub topic and subscription (if no longer needed):
```bash
gcloud pubsub topics delete your-pubsub-topic-id
gcloud pubsub subscriptions delete your-pubsub-subscription-id
```

Congratulations! You have successfully set up a GCP Pub/Sub sink in a Numaflow pipeline.

## Additional Resources

For more detailed information on Numaflow and its usage, visit the [Numaflow Documentation](https://numaflow.numaproj.io/). For GCP Pub/Sub specific configuration and setup, refer to the [Google Cloud Pub/Sub Documentation](https://cloud.google.com/pubsub/docs).
