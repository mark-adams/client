#!/bin/bash

# Deprecated, call build_app.sh directly.

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

PLATFORM=Darwin ./build_app.sh
