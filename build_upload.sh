#!/bin/bash

# Handle the building and uploading of binaries to GitHub releases.

target_platforms="darwin-amd64 linux-amd64 linux-arm"

if [ $# -ne 1 ]; then
  echo "Usage: $0 version_number"
  exit 1
fi

which github-release >/dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "Install github-release first, run: go get github.com/aktau/github-release"
  exit 1
fi

which zip >/dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "zip utility not found, cannot proceed."
  exit 1
fi

github_user=$(jq -e -r .user ~/.github_credentials.json)
github_token=$(jq -e -r .token ~/.github_credentials.json)
if [[ -z "$github_user" || "$github_user" == "null" ]]; then
  echo "Cannot find your GitHub username, do you have a ~/.github_credentials.json file? It should have a key called 'user'."
  exit 1
fi
if [[ -z "$github_token" || "$github_token" == "null" ]]; then
  echo "Cannot find your GitHub token, do you have a ~/.github_credentials.json file? It should have a key called 'token'."
  exit 1
fi

# TODOLATER: validate version# matches semver standards
new_version="$1"
echo "New Version: ${new_version}"
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

set -e

# Adjust version.go
echo "$(date) : Adjusting version.go to reflect new version..."
if [[ $OSTYPE =~ ^darwin ]]; then
  sed -i .bak "s/fs_registrator_version = \".*\"$/fs_registrator_version = \"${new_version}\"/" "${DIR}/version.go"
  rm -f "${DIR}/version.go.bak"
else
  sed -i "s/fs_registrator_version = \".*\"$/fs_registrator_version = \"${new_version}\"/" "${DIR}/version.go"
fi
if [ ! -z "$(git status -s | grep "version.go$")" ]; then
  git pull
  git commit -m "Bump version.go for new release ${new_version}" version.go
  git push
fi
echo "$(date) : Done."

# Remove any old binaries and zips first.
rm -f "${DIR}/bin/*"

# Build for various OSes/archs
echo "$(date) : Building..."
for i in $target_platforms; do
  binary_suffix=${i/-/_}
  echo "$binary_suffix"
  GOOS=$(echo $i | cut -d- -f1) GOARCH=$(echo $i | cut -d- -f2) go build -o "${DIR}/bin/fs-registrator-${new_version}-${binary_suffix}"
done
echo "$(date) : Builds completed"

# Create git tag
echo "$(date) : Tagging and pushing"
git tag "${new_version}"
git push --tags
echo "$(date) : Tagged and pushed"

echo "$(date) : Generate SHA256 hashes of new binaries..."
binary_sha256s="Notes go here.
\`\`\`
"
for i in $target_platforms; do
  binary_suffix=${i/-/_}
  echo "$binary_suffix"
  binary_sha256s="${binary_sha256s}fs-registrator-${new_version}-${binary_suffix}
  $(shasum -a 256 "${DIR}/bin/fs-registrator-${new_version}-${binary_suffix}" | awk '{print $1}')
  
"
done
binary_sha256s="${binary_sha256s}\`\`\`"
echo "$(date) : Hashes generated."

# Create GitHub release
echo "$(date) : Creating release"
github-release release --security-token "$github_token" --user CpuID --repo fs-registrator --tag "${new_version}" --name "fs-registrator ${new_version}" --description "${binary_sha256s}" --pre-release
echo "$(date) : Release created"

# Zip up binaries
echo "$(date) : Zip up binaries for upload to GitHub"
cd "${DIR}/bin"
for i in $target_platforms; do
  binary_suffix=${i/-/_}
  echo "$binary_suffix"
  zip "fs-registrator-${new_version}-${binary_suffix}.zip" "fs-registrator-${new_version}-${binary_suffix}"
done
cd "$DIR"
echo "$(date) : Zip files of binaries created"

# Push binaries up to GitHub
echo "$(date) : Pushing binaries (zipped) to GitHub release"
for i in $target_platforms; do
  binary_suffix=${i/-/_}
  echo "$binary_suffix"
  # Was getting random 502's on these uploads in the past, add a sleep to see if it helps.
  sleep 1
  github-release upload --security-token "$github_token" --user CpuID --repo fs-registrator --tag "${new_version}" --name "fs-registrator-${new_version}-${binary_suffix}.zip" --file "${DIR}/bin/fs-registrator-${new_version}-${binary_suffix}.zip"
done
echo "$(date) : Binaries pushed to GitHub"
