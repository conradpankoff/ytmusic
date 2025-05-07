#!/bin/sh

die() {
  echo $1
  exit 1
}

echo "Running test build..."
go build --tags fts5 -v -o ytmusic || die "Build failed; not syncing files"
echo "Done."

echo "Syncing files..."
rsync \
  --exclude '*.db' \
  --exclude '*.sh' \
  --exclude 'data/*' \
  --exclude 'config.toml' \
  --exclude ytmusic \
  --partial --progress --recursive --times \
  $RSYNC_OPTIONS \
  ./ conrad@conrad-thinkstation-s30:go/src/fknsrs.biz/p/ytmusic/
echo "Done."
