#! /usr/bin/env bash

# This script is the starting point for linux packaging builds. Here's what the
# build does:
#   1) create the "keybase_packaging_v*" docker image, if it doesn't exist
#   2) run the "inside_docker_main.sh" script in that image, sharing several
#      directories from your host, which does all the following...
#   3) build .deb and .rpm packages and lay them out along with the new repo
#      metadata and signatures in your server-ops repo (which is shared
#      read-write with the container)
#   4) commit and push the server-ops repo, or for prerelease builds, push to
#      our S3 bucket
#
# This script mostly concerns itself with updating git repos and organizing
# GPG/SSH/S3 keys for the docker container.

set -e -u -o pipefail

if [ "$#" != 2 ] ; then
  echo Usage: docker_build.sh MODE COMMIT
  exit 1
fi

mode="$1"

here="$(dirname "$BASH_SOURCE")"

clientdir="$(git -C "$here" rev-parse --show-toplevel)"
kbfsdir="$clientdir/../kbfs"
kbfsbetadir="$clientdir/../kbfs-beta"
serveropsdir="$clientdir/../server-ops"

# Run `git fetch` in all the repos we'll share with the container. This
# prevents an unattended build machine from falling behind over time.
for repo in "$clientdir" "$kbfsdir" "$kbfsbetadir" "$serveropsdir" ; do
  echo "Fetching $repo"
  git -C "$repo" fetch
done

# Arrange to share the S3 credentials. We have to do this with a directory
# instead of sharing the file directly, because the latter only works on Linux.
s3cmd_temp="$(mktemp -d)"
cp ~/.s3cfg "$s3cmd_temp"

# Prepare a folder that we'll share with the container, as the container's
# /root directory, where all the build work gets done. Docker recommends that
# write-heavy work happen in shared folders, for better performance.
mkdir -p "$HOME/keybase_builds"
work_dir="$HOME/keybase_builds/$(date +%Y_%m_%d_%H%M%S)_$mode"
mkdir "$work_dir"  # no -p, it's intentionally an error if this exists

# Export the GPG code signing key. We can't just share the ~/.gnupg directory,
# because the host might have a different GnuPG version than the container, and
# GnuPG 2.1 broke back-compat. Sigh. As with S3 above, we need to share the key
# in a directory rather than just a file, for non-Linux support.
code_signing_fingerprint="$(cat "$here/code_signing_fingerprint")"
echo "Exporting the Keybase code signing key ($code_signing_fingerprint)..."
gpg_tempdir="$(mktemp -d)"
gpg_tempfile="$gpg_tempdir/code_signing_key"
gpg --export-secret-key --armor "$code_signing_fingerprint" > "$gpg_tempfile"

# Make sure the Docker image is built.
image=keybase_packaging_v5
if [ -z "$(docker images -q "$image")" ] ; then
  echo "Docker image '$image' not yet built. Building..."
  docker build -t "$image" "$clientdir/packaging/linux"
fi

echo '=== docker ==='
docker run \
  -v "$work_dir:/root" \
  -v "$clientdir:/CLIENT:ro" \
  -v "$kbfsdir:/KBFS:ro" \
  -v "$kbfsbetadir:/KBFSBETA:ro" \
  -v "$serveropsdir:/SERVEROPS:ro" \
  -v "$gpg_tempdir:/GPG" \
  -v "$HOME/.ssh:/SSH:ro" \
  -v "$s3cmd_temp:/S3CMD:ro" \
  -e BUCKET_NAME \
  "$image" \
  bash /CLIENT/packaging/linux/inside_docker_main.sh "$@"
