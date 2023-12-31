#!/bin/bash

# This file updates the vmlinux.h file in this directory to match the current
# kernel. This is OK to ship for many different kernel versions because we use
# CO:RE.
#
# Depends on bpftool compiled against the current kernel.

set -euo pipefail
cd "$(dirname "$0")"

vmlinux="./$(uname -m)/vmlinux.h"

# Attach the license header and source URL.
cat <<EOF > "$vmlinux"
// This file was generated on the following system:
//   $(uname -a)
// On $(date).
//
// Kernel headers licensed under GPL-2.0.

EOF

bpftool btf dump file /sys/kernel/btf/vmlinux format c >> "$vmlinux"

# Remove extra trailing newlines.
sed -i -e :a -e '/^\n*$/{$d;N;};/\n$/ba' "$vmlinux"
