module github.com/netobserv/network-observability-operator

go 1.22.3
toolchain go1.23.4

require (
	github.com/bpfman/bpfman-operator v0.5.5-0.20241023163832-0bf84bbd3927
	github.com/coreos/go-semver v0.3.1
	github.com/go-logr/logr v1.4.2
	github.com/google/go-cmp v0.6.0
	github.com/netobserv/flowlogs-pipeline v1.7.0-community.0.20241126155124-18e017f67001
	github.com/netobserv/netobserv-ebpf-agent v1.6.1-crc2.0.20241008130234-a20397fb8f88
	github.com/onsi/ginkgo/v2 v2.22.0
	github.com/onsi/gomega v1.35.1
	github.com/openshift/api v0.0.0-20240722135205-ae4f370f361f
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.78.2
	github.com/prometheus/common v0.55.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.9.0
	go.uber.org/zap v1.27.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.31.3
	k8s.io/apiextensions-apiserver v0.31.2
	k8s.io/apimachinery v0.31.3
	k8s.io/client-go v0.31.3
	k8s.io/klog v1.0.0
	k8s.io/kube-aggregator v0.31.2
	k8s.io/utils v0.0.0-20240902221715-702e33fdd3c3
	sigs.k8s.io/controller-runtime v0.19.3
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/agoda-com/opentelemetry-logs-go v0.5.0 // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/hub v1.0.1 // indirect
	github.com/cenkalti/rpc2 v0.0.0-20210604223624-c1acbc6ec984 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cilium/ebpf v0.16.0 // indirect
	github.com/containernetworking/cni v1.1.2 // indirect
	github.com/containernetworking/plugins v1.2.0 // indirect
	github.com/coreos/go-iptables v0.6.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emicklei/go-restful/v3 v3.12.1 // indirect
	github.com/evanphx/json-patch v5.9.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/gavv/monotime v0.0.0-20190418164738-30dba4353424 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-kit/kit v0.13.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20241029153458-d1b30febd7db // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gopacket/gopacket v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20211123025425-613501dd5deb // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/ip2location/ip2location-go/v9 v9.7.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/libp2p/go-reuseport v0.3.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mdlayher/ethernet v0.0.0-20220221185849-529eae5b6118 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/minio-go/v7 v7.0.77 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/netobserv/gopipes v0.3.0 // indirect
	github.com/netobserv/loki-client-go v0.0.0-20220927092034-f37122a54500 // indirect
	github.com/netsampler/goflow2 v1.3.7 // indirect
	github.com/ovn-org/libovsdb v0.7.1-0.20240820095311-ce1951614a20 // indirect
	github.com/ovn-org/ovn-kubernetes/go-controller v0.0.0-20241126140656-c95491e46334 // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/pion/dtls/v2 v2.2.4 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/transport/v2 v2.0.0 // indirect
	github.com/pion/udp v0.1.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/prometheus/prometheus v1.8.2-0.20201028100903-3245b3267b24 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/safchain/ethtool v0.3.1-0.20231027162144-83e5e0097c91 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/urfave/cli/v2 v2.27.2 // indirect
	github.com/vishvananda/netlink v1.3.0 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/vmware/go-ipfix v0.9.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xrash/smetrics v0.0.0-20240312152122-5f08fbb34913 // indirect
	go.opentelemetry.io/otel v1.31.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.29.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.29.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.28.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.31.0 // indirect
	go.opentelemetry.io/otel/sdk v1.29.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.29.0 // indirect
	go.opentelemetry.io/otel/trace v1.31.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/oauth2 v0.22.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/term v0.25.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/time v0.6.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240822170219-fc7c04adadcd // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240822170219-fc7c04adadcd // indirect
	google.golang.org/grpc v1.67.1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/gcfg.v1 v1.2.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20240808142205-8e686545bdb8 // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)

replace github.com/prometheus/common v0.55.0 => github.com/netobserv/prometheus-common v0.55.0-netobserv

replace github.com/netobserv/netobserv-ebpf-agent => github.com/msherif1234/netobserv-ebpf-agent v0.0.0-20241213142900-5c07db3a6c0b
