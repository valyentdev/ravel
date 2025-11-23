_This repo is now archived._

# Ravel

> Ravel is an open-source containers-as-microVMs orchestrator.

<div align="center">
  <a href="https://discord.gg/HXkpCG7DEH">
    <img src="https://img.shields.io/badge/chat-on%20discord-7289DA.svg" alt="Discord Chat" />
  </a>

  <a href="https://x.com/intent/follow?screen_name=ValyentCloud">
    <img src="https://img.shields.io/twitter/follow/valyentcloud.svg?label=Follow%20@valyentcloud" alt="Follow @valyentcloud" />
  </a>
</div>

## Table of Contents

- [Technologies](#technologies)
- [Roadmap](#roadmap)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Documentation](#documentation)
- [FAQ](#faq)
- [License](#license)
- [Star History](#star-history)

## About

Ravel emerges as the building block for [Valyent](https://valyent.cloud)'s cloud services.

Ravel is an open-source containers-as-microVMs as _microVMs_ orchestrator. It allows you to create, manage, and destroy microVMs on the fly. It supports running OCI images inside microVMs powered by CloudHypervisor.

## Technologies

- [Go](https://golang.org/): A fast, efficient programming language designed for building scalable software.
- [Cloud Hypervisor](https://github.com/cloud-hypervisor/cloud-hypervisor): A lightweight virtual machine monitor for running modern cloud workloads.
- [NATS](https://nats.io/): For publish/subscribe features
- [Corrosion](https://github.com/superfly/corrosion): Gossip-based service discovery (and more) for large distributed systems.
- [Containerd](https://containerd.io/): For image management

## Features

- [x] Run OCI images inside cloud-hypervisor micro-VMs with Ravel Runtime
- [x] An intuitive API to and manage Ravel machines
- [x] Mutual TLS cluster-communication
- [x] An HTTP with TLS
- [ ] Volumes management (work in progress)
- [ ] Secrets management (coming soon)
- [ ] Wireguard-based private networks (coming soon)

## Prerequisites

- [Go 1.22](https://golang.org/dl/)
- [Cloud Hypervisor](https://github.com/cloud-hypervisor/cloud-hypervisor)
- [TUN kernel module](https://en.wikipedia.org/wiki/TUN/TAP) enabled
- [KVM](https://fr.wikipedia.org/wiki/Kernel-based_Virtual_Machine) enabled

## Documentation

To try out Ravel features, you can look at our [documentation](https://docs.valyent.cloud/installation). To install Ravel, you can follow the [Ravel documentation](./docs/README.md).

## FAQ

### Is it production-ready?

Not yet, Ravel is in **alpha** and is to be considered unstable.

But we are working full time on providing a stable release.

### Why is it named Ravel?

Ravel is named after the famous composer Maurice Ravel, known for his orchestral works.

### How do I contribute?

Please come and join us on our [Discord server](https://discord.valyent.cloud), where you can ask questions, get help, and contribute to the project.

### How do I report a bug?

Please open an issue on our [GitHub repository](https://github.com/valyentdev/ravel/issues).

### How do I request a feature?

Please open an issue on our [GitHub repository](https://github.com/valyentdev/ravel/issues).

## License

Copyright 2024 - Valyent

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Star History

Thank you for your support! 🌟

[![Star History Chart](https://api.star-history.com/svg?repos=valyentdev/ravel&type=Date)](https://star-history.com/#valyentdev/ravel&Date)
