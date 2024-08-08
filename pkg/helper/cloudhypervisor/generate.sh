#!/bin/bash

SPEC_URL=https://raw.githubusercontent.com/cloud-hypervisor/cloud-hypervisor/main/vmm/src/api/openapi/cloud-hypervisor.yaml

oapi-codegen -config ./oapi.yaml $SPEC_URL
