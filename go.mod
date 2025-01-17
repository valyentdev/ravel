module github.com/valyentdev/ravel

go 1.23.2

replace github.com/valyentdev/ravel/api => ./api

require (
	github.com/c9s/goprocinfo v0.0.0-20210130143923-c95fcf8c64a8
	github.com/cloudflare/cfssl v1.6.5
	github.com/containerd/cgroups/v3 v3.0.5
	github.com/containerd/containerd/v2 v2.0.2
	github.com/containerd/errdefs v1.0.0
	github.com/coreos/go-iptables v0.8.0
	github.com/danielgtaylor/huma/v2 v2.27.0
	github.com/fsnotify/fsnotify v1.8.0
	github.com/gammazero/deque v1.0.0
	github.com/google/go-containerregistry v0.20.2
	github.com/jackc/pgx/v5 v5.7.2
	github.com/jackc/tern/v2 v2.3.2
	github.com/mistifyio/go-zfs/v3 v3.0.1
	github.com/nats-io/nats.go v1.38.0
	github.com/nrednav/cuid2 v1.0.1
	github.com/oklog/ulid v1.3.1
	github.com/opencontainers/image-spec v1.1.0
	github.com/pelletier/go-toml/v2 v2.2.3
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.10.0
	github.com/u-root/u-root v0.14.0
	github.com/valyentdev/corroclient v0.2.2
	go.etcd.io/bbolt v1.3.11
	sigs.k8s.io/yaml v1.4.0

)

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20240806141605-e8a1dd7889d6 // indirect
	github.com/AdamKorcz/go-118-fuzz-build v0.0.0-20231105174938-2b5cbb29f3e2 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hcsshim v0.12.9 // indirect
	github.com/cilium/ebpf v0.17.1 // indirect
	github.com/containerd/containerd/api v1.8.0 // indirect
	github.com/containerd/continuity v0.4.5 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/fifo v1.1.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v1.0.0-rc.1 // indirect
	github.com/containerd/plugin v1.0.0 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.16.3 // indirect
	github.com/containerd/ttrpc v1.2.7 // indirect
	github.com/containerd/typeurl/v2 v2.2.3 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/cli v27.5.0+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.2 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/certificate-transparency-go v1.3.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/mdlayher/socket v0.5.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/signal v0.7.1 // indirect
	github.com/moby/sys/user v0.3.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/opencontainers/runtime-spec v1.2.0 // indirect
	github.com/opencontainers/runtime-tools v0.9.1-0.20221107090550-2e043c6bd626 // indirect
	github.com/opencontainers/selinux v1.11.1 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	github.com/u-root/uio v0.0.0-20240224005618-d2acac8f3701 // indirect
	github.com/vbatts/tar-split v0.11.6 // indirect
	github.com/weppos/publicsuffix-go v0.40.3-0.20241218111332-1518a6f1cb34 // indirect
	github.com/zmap/zcrypto v0.0.0-20241218205341-e47c4f4da3a4 // indirect
	github.com/zmap/zlint/v3 v3.6.5 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0 // indirect
	go.opentelemetry.io/otel v1.33.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	golang.org/x/mod v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250106144421-5f5ef82da422 // indirect
	google.golang.org/grpc v1.69.4 // indirect
	google.golang.org/protobuf v1.36.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	tags.cncf.io/container-device-interface v0.8.0 // indirect
	tags.cncf.io/container-device-interface/specs-go v0.8.0 // indirect
)

require (
	github.com/mdlayher/vsock v1.2.1
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.4-0.20230606125235-dd1b4c2e81af
	github.com/valyentdev/ravel/api v0.0.0-20250115105712-986972f8fc94
	github.com/vishvananda/netlink v1.3.0
	github.com/vishvananda/netns v0.0.5
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.34.0
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0
	golang.org/x/text v0.21.0 // indirect
)
