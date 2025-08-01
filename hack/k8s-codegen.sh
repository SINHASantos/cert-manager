#!/usr/bin/env bash

# Copyright 2022 The cert-manager Authors.
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

clientgen=$1
deepcopygen=$2
informergen=$3
listergen=$4
defaultergen=$5
conversiongen=$6
openapigen=$7
applyconfigurationgen=$8

echo "+++ Generating code..." >&2

module_name="github.com/cert-manager/cert-manager"

# Generate deepcopy functions for all internal and external APIs
deepcopy_inputs=(
  pkg/apis/certmanager/v1 \
  internal/apis/certmanager \
  pkg/apis/acme/v1 \
  internal/apis/acme \
  pkg/apis/config/cainjector/v1alpha1 \
  internal/apis/config/cainjector \
  pkg/apis/config/webhook/v1alpha1 \
  internal/apis/config/webhook \
  pkg/apis/config/controller/v1alpha1 \
  internal/apis/config/controller \
  pkg/apis/config/shared/v1alpha1 \
  internal/apis/config/shared \
  pkg/apis/meta/v1 \
  internal/apis/meta \
  pkg/acme/webhook/apis/acme/v1alpha1 \
)

# Used for generating apply configurations and client openapi specs.
# Separate to client_inputs because we need apply configurations for metav1,
# and client-gen has no way to exclude a input package just using markers in code.
api_inputs=(
  pkg/apis/certmanager/v1 \
  pkg/apis/acme/v1 \
  pkg/apis/meta/v1 \
)

client_subpackage="pkg/client"
client_package="${module_name}/${client_subpackage}"
# Generate clientsets, listers and informers for user-facing API types
client_inputs=(
  pkg/apis/certmanager/v1 \
  pkg/apis/acme/v1 \
)

# Generate defaulting functions to be used by the mutating webhook
defaulter_inputs=(
  internal/apis/certmanager/v1 \
  internal/apis/acme/v1 \
  internal/apis/config/shared/v1alpha1 \
  internal/apis/config/cainjector/v1alpha1 \
  internal/apis/config/webhook/v1alpha1 \
  internal/apis/config/controller/v1alpha1 \
  internal/apis/meta/v1 \
)

# Generate conversion functions to be used by the conversion webhook
conversion_inputs=(
  internal/apis/certmanager/v1 \
  internal/apis/acme/v1 \
  internal/apis/config/shared/v1alpha1 \
  internal/apis/config/cainjector/v1alpha1 \
  internal/apis/config/webhook/v1alpha1 \
  internal/apis/config/controller/v1alpha1 \
  internal/apis/meta/v1 \
)

# clean will delete files matching name in path.
clean() {
  path=$1
  name=$2
  if [[ ! -d "$path" ]]; then
    return 0
  fi
  find "$path" -name "$name" -delete
}

gen-openapi-acme() {
  clean pkg/acme/webhook/openapi 'zz_generated.openapi.go'
  echo "+++ Generating ACME openapi..." >&2
  mkdir -p hack/openapi_reports
  "$openapigen" \
    --go-header-file "hack/boilerplate-go.txt" \
    --report-filename "hack/openapi_reports/acme.txt" \
    --output-dir ./pkg/acme/webhook/openapi/ \
    --output-pkg "github.com/cert-manager/cert-manager/pkg/acme/webhook/openapi" \
		--output-file zz_generated.openapi.go \
    "k8s.io/apimachinery/pkg/version" \
    "k8s.io/apimachinery/pkg/runtime" \
    "k8s.io/apimachinery/pkg/apis/meta/v1" \
    "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1" \
    "github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
}

gen-openapi-client() {
  clean internal/generated/openapi 'zz_generated.openapi.go'
  echo "+++ Generating client openapi..." >&2
  prefixed_inputs=( "${api_inputs[@]/#/$module_name/}" )
  "$openapigen" \
    --go-header-file "hack/boilerplate-go.txt" \
    --report-filename "hack/openapi_reports/client.txt" \
    --output-dir ./internal/generated/openapi/ \
    --output-pkg "github.com/cert-manager/cert-manager/internal/generated/openapi" \
		--output-file zz_generated.openapi.go \
		"k8s.io/api/core/v1" \
    "k8s.io/apimachinery/pkg/version" \
    "k8s.io/apimachinery/pkg/runtime" \
    "k8s.io/apimachinery/pkg/apis/meta/v1" \
    "k8s.io/apimachinery/pkg/api/resource" \
    "k8s.io/apimachinery/pkg/util/intstr" \
    "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1" \
    "k8s.io/component-base/logs/api/v1" \
    "sigs.k8s.io/gateway-api/apis/v1" \
    "${prefixed_inputs[@]}"
}

gen-deepcopy() {
  clean pkg/apis 'zz_generated.deepcopy.go'
  clean pkg/acme/webhook/apis 'zz_generated.deepcopy.go'
  clean pkg/webhook/handlers/testdata/apis 'zz_generated.deepcopy.go'
  echo "+++ Generating deepcopy methods..." >&2
  prefixed_inputs=( "${deepcopy_inputs[@]/#/$module_name/}" )
  "$deepcopygen" \
    --go-header-file hack/boilerplate-go.txt \
    --output-file zz_generated.deepcopy.go \
    --bounding-dirs "${module_name}" \
    "${prefixed_inputs[@]}"
}

gen-applyconfigurations() {
  # This is a temporary hack to generate the schema YAMLs
  # required to generate fake clientsets that actually works.
  # Upstream issue: https://github.com/kubernetes/kubernetes/issues/126850
  GOPROXY=off go install \
    "${module_name}/internal/generated/openapi/cmd/models-schema"

  clean "${client_subpackage}"/applyconfigurations '*.go'
  echo "+++ Generating applyconfigurations..." >&2
  prefixed_inputs=( "${api_inputs[@]/#/$module_name/}" )
 "$applyconfigurationgen" \
    --go-header-file hack/boilerplate-go.txt \
    --openapi-schema <($(go env GOPATH)/bin/models-schema) \
    --output-dir "${client_subpackage}"/applyconfigurations \
    --output-pkg "${client_package}"/applyconfigurations \
    "${prefixed_inputs[@]}"
}

gen-clientsets() {
  clean "${client_subpackage}"/clientset '*.go'
  echo "+++ Generating clientsets..." >&2
  prefixed_inputs=( "${client_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "$clientgen" \
    --go-header-file hack/boilerplate-go.txt \
    --clientset-name versioned \
    --apply-configuration-package "${client_package}"/applyconfigurations \
    --input-base "" \
    --input "$joined" \
    --output-dir "${client_subpackage}"/clientset \
    --output-pkg "${client_package}"/clientset
}

gen-listers() {
  clean "${client_subpackage}/listers" '*.go'
  echo "+++ Generating listers..." >&2
  prefixed_inputs=( "${client_inputs[@]/#/$module_name/}" )
  "$listergen" \
    --go-header-file hack/boilerplate-go.txt \
    --output-dir "${client_subpackage}"/listers \
    --output-pkg "${client_package}"/listers \
    "${prefixed_inputs[@]}"
}

gen-informers() {
  clean "${client_subpackage}"/informers '*.go'
  echo "+++ Generating informers..." >&2
  prefixed_inputs=( "${client_inputs[@]/#/$module_name/}" )
  "$informergen" \
    --go-header-file hack/boilerplate-go.txt \
    --versioned-clientset-package "${client_package}"/clientset/versioned \
    --listers-package "${client_package}"/listers \
    --output-dir "${client_subpackage}"/informers \
    --output-pkg "${client_package}"/informers \
    "${prefixed_inputs[@]}"
}

gen-defaulters() {
  clean internal/apis 'zz_generated.defaults.go'
  clean pkg/webhook/handlers/testdata/apis 'zz_generated.defaults.go'
  echo "+++ Generating defaulting functions..." >&2
  
  DEFAULT_EXTRA_PEER_PKGS=(
    github.com/cert-manager/cert-manager/internal/apis/meta \
    github.com/cert-manager/cert-manager/internal/apis/meta/v1 \
    github.com/cert-manager/cert-manager/internal/apis/config/shared \
    github.com/cert-manager/cert-manager/internal/apis/config/shared/v1alpha1 \
    github.com/cert-manager/cert-manager/pkg/apis/meta/v1 \
    github.com/cert-manager/cert-manager/pkg/apis/config/shared/v1alpha1 \
  )
  DEFAULT_PKGS=( "${defaulter_inputs[@]/#/$module_name/}" )

  "$defaultergen" \
    --go-header-file hack/boilerplate-go.txt \
    --extra-peer-dirs "$( IFS=$','; echo "${DEFAULT_EXTRA_PEER_PKGS[*]}" )" \
    --output-file zz_generated.defaults.go \
    "${DEFAULT_PKGS[@]}"
}

gen-conversions() {
  clean internal/apis 'zz_generated.conversion.go'
  clean pkg/webhook/handlers/testdata/apis 'zz_generated.conversion.go'
  echo "+++ Generating conversion functions..." >&2

  CONVERSION_EXTRA_PEER_PKGS=(
    github.com/cert-manager/cert-manager/internal/apis/meta \
    github.com/cert-manager/cert-manager/internal/apis/meta/v1 \
    github.com/cert-manager/cert-manager/internal/apis/config/shared \
    github.com/cert-manager/cert-manager/internal/apis/config/shared/v1alpha1 \
    github.com/cert-manager/cert-manager/pkg/apis/meta/v1 \
    github.com/cert-manager/cert-manager/pkg/apis/config/shared/v1alpha1 \
  )
  CONVERSION_PKGS=( "${conversion_inputs[@]/#/$module_name/}" )

  "$conversiongen" \
      --go-header-file hack/boilerplate-go.txt \
      --extra-peer-dirs "$( IFS=$','; echo "${CONVERSION_EXTRA_PEER_PKGS[*]}" )" \
      --output-file zz_generated.conversion.go \
      "${CONVERSION_PKGS[@]}"
}

gen-openapi-acme
gen-openapi-client
gen-deepcopy
gen-applyconfigurations
gen-clientsets
gen-listers
gen-informers
gen-defaulters
gen-conversions
