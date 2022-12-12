# Network Observability Architecture

The Network Observability solution consists on a [Network Observability Operator (NOO)](https://github.com/netobserv/network-observability-operator)
that deploys, configures and controls the status of the following components:

* [Network Observability eBPF Agent](https://github.com/netobserv/netobserv-ebpf-agent/)
  * It is attached to all the interfaces in the host network and listen for each network packet that
    is submitted or received by their egress/ingress. The agent aggregates all the packets by source
    and destination addresses, protocol, etc... into network flows that are submitted to the
    Flowlogs-Pipeline flow processor.
* [Network Observabiilty Flowlogs-Pipeline (FLP)](https://github.com/netobserv/flowlogs-pipeline)
  * It receives the raw flows from the agent and decorates them with Kubernetes information (Pod
    and host names, namespaces, etc.), and stores them as JSON into a [Loki](https://grafana.com/oss/loki/)
    instance.
* [Network Observability Console Plugin](https://github.com/netobserv/network-observability-console-plugin)
  * It is attached to the Openshift console as a plugin (see Figure 1, though it can be also
    deployed in standalone mode). The Console Plugin queries the flows information stored in Loki
    and allows filtering flows, showing network topologies, etc.

![Netobserv frontend architecture](./assets/frontend.png)
Figure 1: Console Plugin deployment

There are two existing deployment modes for Network Observability: direct mode and Kafka mode.

## Direct-mode deployment

In direct mode (figure 2), the eBPF agent sends the flows information to Flowlogs-Pipeline encoded as Protocol
Buffers (binary representation) via [gRPC](https://grpc.io/). In this scenario, Flowlogs-Pipeline
is usually deployed as a DaemonSet so there is a 1:1 communication between the Agent and FLP internal
to the host, so we minimize cluster network usage.

![Netobserv component's architecture (direct mode)](./assets/architecture-direct.png)
Figure 2: Direct deployment

## Kafka-mode deployment

In Kafka mode (figure 3), the communication between the eBFP agent and FLP is done via a Kafka topic.

![Netobserv component's architecture (Kafka mode)](./assets/architecture-kafka.png)
Figure 3: Kafka deployment

This has some advantages over the direct mode:
1. The flows' are buffered in the Kafka topic, so if there is a peak of flows, we make sure that
   FLP will receive/process them without any kind of denial of service.
2. Flows are persisted in the topic, so if FLP is restarted by any reason (an update in the
   configuration or just a crash), the forwarded flows are persisted in Kafka for its later
   processing, and we don't lose them.
3. Deploying FLP as a deployment, you don't have to keep the 1:1 proportion. You can scale up and
   down FLP pods according to your load.