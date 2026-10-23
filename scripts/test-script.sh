#!/usr/bin/env bash

set -e

variant=${1:-${VARIANT}}
deployment=${2:-${DEPLOYMENT_NAME}}
namespace=${3:-${NAMESPACE}}

[[ -z $namespace ]] && namespace="default"

timed() {
  end=$(date +%s)
  elapsed=$(($end - $1))
  printf "\nTotal runtime: %02d min %02d seconds\n" "$((elapsed / 60))" "$((elapsed % 60))"
}

run_folder() {
  newman run --delay-request=100 --folder="$1" \
    --export-environment "$variant"/postman/environment.json \
    --environment "$variant"/postman/environment.json \
    "$variant"/postman/collection.json
}

step() {
  local number=$1
  local replicas=0
  [[ $((number % 2)) -eq 0 ]] && replicas=1

  printf "=== Step %d: scale %s to %s ===\n" "$number" "$deployment" "$replicas"
  kubectl scale deployment "$deployment" -n "$namespace" --replicas="$replicas"
  if [[ "$replicas" -eq 1 ]]; then
    kubectl rollout status deployment "$deployment" -n "$namespace" --timeout=120s
  fi
  run_folder "step$number"
}

start=$(date +%s)
trap 'timed $start' EXIT

run_folder success
step 1
step 2
step 3
step 4
