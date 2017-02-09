#!/bin/bash

# Copyright 2014 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

# This script is only used for e2e tests! Don't use it in production!
# This is also a temporary bridge to slowly switch over everything to
# federation/develop.sh. Carefully moving things step-by-step, ensuring
# things don't break.
# TODO(madhusudancs): Remove this script and its dependencies.


KUBE_ROOT=$(dirname "${BASH_SOURCE}")/../..
# For `kube::log::status` function since it already sources
# "${KUBE_ROOT}/cluster/lib/logging.sh" and DEFAULT_KUBECONFIG
source "${KUBE_ROOT}/cluster/common.sh"
# For $FEDERATION_PUSH_REPO_BASE and $FEDERATION_NAMESPACE.
source "${KUBE_ROOT}/federation/cluster/common.sh"

FEDERATION_NAME="${FEDERATION_NAME:-e2e-federation}"
FEDERATION_KUBE_CONTEXT="${FEDERATION_KUBE_CONTEXT:-e2e-federation}"
DNS_ZONE_NAME="${FEDERATION_DNS_ZONE_NAME:-}"
HOST_CLUSTER_CONTEXT="${FEDERATION_HOST_CLUSTER_CONTEXT:-${1}}"

# get_version returns the version in KUBERNETES_RELEASE or defaults to the
# value in the federation `versions` file.
# TODO(madhusudancs): This is a duplicate of the function in
# federation/develop/develop.sh with a minor difference. This
# function tries to default to the version information in
# _output/federation/versions file where as the one in develop.sh
# tries to default to the version in the kubernetes versions file.
# These functions should be consolidated to read the version from
# kubernetes version defs file.
function get_version() {
  local -r versions_file="${KUBE_ROOT}/_output/federation/versions"

  if [[ -n "${KUBERNETES_RELEASE:-}" ]]; then
    echo "${KUBERNETES_RELEASE//+/_}"
    return
  fi

  if [[ ! -f "${versions_file}" ]]; then
    echo "Couldn't determine the release version: neither the " \
     "KUBERNETES_RELEASE environment variable is set, nor does " \
     "the versions file exist at ${versions_file}"
    exit 1
  fi

  # Read the version back from the versions file if no version is given.
  local -r kube_version="$(cat "${versions_file}" | python -c '\
import json, sys;\
print json.load(sys.stdin)["KUBE_VERSION"]')"

  echo "${kube_version//+/_}"
}

# Initializes the control plane.
# TODO(madhusudancs): Move this to federation/develop.sh.
function init() {
  kube::log::status "Deploying federation control plane for ${FEDERATION_NAME} in cluster ${HOST_CLUSTER_CONTEXT}"

  local -r project="${KUBE_PROJECT:-${PROJECT:-}}"
  local -r kube_registry="${KUBE_REGISTRY:-gcr.io/${project}}"
  local -r kube_version="$(get_version)"

  "${KUBE_ROOT}/federation/develop/kubefed.sh" init \
      "${FEDERATION_NAME}" \
      --host-cluster-context="${HOST_CLUSTER_CONTEXT}" \
      --dns-zone-name="${DNS_ZONE_NAME}" \
      --image="${kube_registry}/hyperkube-amd64:${kube_version}"
}

# join_cluster_to_federation joins the clusters in the local kubeconfig to federation. The clusters
# and their kubeconfig entries in the local kubeconfig are created while deploying clusters, i.e. when kube-up is run.
function join_cluster_to_federation() {
  for cluster in $("${KUBE_ROOT}/cluster/kubectl.sh" config get-clusters |sed -n '1!p'); do
    # Skip federation context
    if [[ "${cluster}" == "${FEDERATION_NAME}" ]]; then
      continue
    fi
    # Skip contexts not beginning with "federation"
    if [[ "${cluster}" != federation* ]]; then
      continue
    fi

  kube::log::status "Joining cluster with name '${cluster}' to federation with name '${FEDERATION_NAME}'"

  "${KUBE_ROOT}/federation/develop/kubefed.sh" join \
      "${cluster}" \
      --host-cluster-context="${HOST_CLUSTER_CONTEXT}" \
      --context="${FEDERATION_NAME}" \
      --secret-name="${cluster//_/-}"    # Replace "_" by "-"
  done
}

USE_KUBEFED="${USE_KUBEFED:-}"

if [[ "${USE_KUBEFED}" == "true" ]]; then
  init

  join_cluster_to_federation
else
  export FEDERATION_IMAGE_TAG="$(get_version)"
  create-federation-api-objects
fi
