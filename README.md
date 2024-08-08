<div align="center">
  <picture>
    <img src="./logo.png" alt="Logo" width="150" height="auto" />
  </picture>
</div>
<div align="center">
  <a href="https://discord.gg/DuW5uQCtZj">
    <img src="https://dcbadge.vercel.app/api/server/DuW5uQCtZj)](https://discord.gg/DuW5uQCtZj">
  </a>
  <a href="https://x.com/valyentdev">
    <img src="https://img.shields.io/badge/X-%23000000.svg?style=for-the-badge&logo=X&logoColor=white">
  </a>
</div>
<h1 align="center">
  Ravel ♪
</h1>

> Ravel is an open-source microVMs orchestrator.

> [!WARNING]
>
> Ravel is in **ALPHA**, and is to be considered unstable.
>
> We are working on a stable release.

## Table of Contents

- [Technologies](#technologies)
- [Roadmap] (#roadmap)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Documentation](#documentation)
- [FAQ](#faq)
- [License](#license)
- [Star History](#star-history)

## WIP Roadmap

There is still a lot of work to be done to stabilize Ravel and make it ready for production:

- [ ] Create a process jailer for cloud-hypervisor (see the firecracker Jailer)
- [ ] Enforce resources limitations with cgroups
- [ ] Secure agent / manager communications with mTLS
- [ ] Improve the OCI images management on the agent
- [ ] Private networks with Wireguard
- [ ] Build a solid implementation of the server to ensure consistency and authorization
- [ ] Support public IPV6 address for machines
- [ ] Build a service discovery system and a proxy around Corrosion
- [ ] Implement machine migrations between workers
- [ ] Persistent volumes
- [ ] Lot of tests
- [ ] Handle different regions properly

## About

Ravel emerges as the building block for [Valyent](https://valyent.dev)'s cloud services.

Ravel is a **bidding-style orchestrator** for _microVMs_. It allows you to create, manage, and destroy microVMs on the fly. It supports running OCI images inside cloud-hypervisor micro-vms.

## Technologies

- [Go](https://golang.org/): A fast, efficient programming language designed for building scalable software.
- [Cloud Hypervisor](https://github.com/cloud-hypervisor/cloud-hypervisor): A lightweight virtual machine monitor for running modern cloud workloads.
- [NATS](https://nats.io/): A simple, high-performance messaging system for cloud applications and microservices.
- [Corrosion](https://github.com/superfly/corrosion): Gossip-based service discovery (and more) for large distributed systems.

## Features

- [x] Create, manage and destroy microVMs
- [x] RESTful API
- [x] Bidding-style orchestrator
- [ ] Private networks
- [ ] Secrets management
- [ ] Multi-region
- [ ] Multi-tenancy

## Prerequisites

- [Go 1.22](https://golang.org/dl/)
- [Cloud Hypervisor](https://github.com/cloud-hypervisor/cloud-hypervisor)
- [TUN kernel module](https://en.wikipedia.org/wiki/TUN/TAP) enabled
- [KVM](https://fr.wikipedia.org/wiki/Kernel-based_Virtual_Machine) enabled

## Documentation

For more details, please refer to our [documentation](https://ravel.sh).

## FAQ

### Why is it named Ravel?

Ravel is named after the famous composer Maurice Ravel, known for his orchestral works.

### How do I contribute?

Please come and join us on our [Discord server](https://discord.gg/DuW5uQCtZj), where you can ask questions, get help, and contribute to the project.

### How do I report a bug?

Please open an issue on our [GitHub repository](https://github.com/valyentdev/ravel/issues).

### How do I request a feature?

Please open an issue on our [GitHub repository](https://github.com/valyentdev/ravel/issues).

### Is it production-ready?

No, Ravel is in **ALPHA** and is to be considered unstable.

## License

Copyright 2024 - Valyent

This project is licensed under the X License - see the [license](./LICENSE.md) file for details.

## Star History

Thank you for your support! 🌟

[![Star History Chart](https://api.star-history.com/svg?repos=valyentdev/ravel&type=Date)](https://star-history.com/#valyentdev/ravel&Date)
