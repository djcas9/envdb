#!/bin/sh
set -e

NAME="envdb"
BRANCH="master"

VERSION=$(cat .Version)
OLD_VERSION="0.2.2"
YANK_OLD_VERSIONS=true

PREFIX="mephux/$NAME"
RELEASE_PATH="release"

DEB="${NAME}_${OLD_VERSION}_amd64.deb"
DEB386="${NAME}_${OLD_VERSION}_386.deb"

if [ "$DRONE_BRANCH" = "$BRANCH" ] && [ "$DRONE_PR" != "true" ]; then
  echo "MASTER BRANCH: Deploying..."
  make release

  if [ "$YANK_OLD_VERSIONS" = true ]; then

    package_cloud yank $PREFIX/ubuntu/lucid   $DEB
    package_cloud yank $PREFIX/ubuntu/hardy   $DEB
    package_cloud yank $PREFIX/ubuntu/utopic  $DEB
    package_cloud yank $PREFIX/ubuntu/precise $DEB
    package_cloud yank $PREFIX/ubuntu/trusty  $DEB

    package_cloud yank $PREFIX/ubuntu/lucid   $DEB386
    package_cloud yank $PREFIX/ubuntu/hardy   $DEB386
    package_cloud yank $PREFIX/ubuntu/utopic  $DEB386
    package_cloud yank $PREFIX/ubuntu/precise $DEB386
    package_cloud yank $PREFIX/ubuntu/trusty  $DEB386

    package_cloud yank $PREFIX/debian/squeeze $DEB
    package_cloud yank $PREFIX/debian/jessie  $DEB
    package_cloud yank $PREFIX/debian/wheezy  $DEB

    package_cloud yank $PREFIX/debian/squeeze $DEB386
    package_cloud yank $PREFIX/debian/jessie  $DEB386
    package_cloud yank $PREFIX/debian/wheezy  $DEB386

  fi

  #
  # Push New Packages
  #

  # Ubuntu - Push - amd64
  package_cloud push $PREFIX/ubuntu/lucid   $RELEASE_PATH/$NAME-amd64.deb
  package_cloud push $PREFIX/ubuntu/hardy   $RELEASE_PATH/$NAME-amd64.deb
  package_cloud push $PREFIX/ubuntu/utopic  $RELEASE_PATH/$NAME-amd64.deb
  package_cloud push $PREFIX/ubuntu/precise $RELEASE_PATH/$NAME-amd64.deb
  package_cloud push $PREFIX/ubuntu/trusty  $RELEASE_PATH/$NAME-amd64.deb

  # Ubuntu - Push - 386
  package_cloud push $PREFIX/ubuntu/lucid   $RELEASE_PATH/$NAME-386.deb
  package_cloud push $PREFIX/ubuntu/hardy   $RELEASE_PATH/$NAME-386.deb
  package_cloud push $PREFIX/ubuntu/utopic  $RELEASE_PATH/$NAME-386.deb
  package_cloud push $PREFIX/ubuntu/precise $RELEASE_PATH/$NAME-386.deb
  package_cloud push $PREFIX/ubuntu/trusty  $RELEASE_PATH/$NAME-386.deb

  # Debian - Push - amd64
  package_cloud push $PREFIX/debian/squeeze $RELEASE_PATH/$NAME-amd64.deb
  package_cloud push $PREFIX/debian/jessie  $RELEASE_PATH/$NAME-amd64.deb
  package_cloud push $PREFIX/debian/wheezy  $RELEASE_PATH/$NAME-amd64.deb

  # Debian - Push - 386
  package_cloud push $PREFIX/debian/squeeze $RELEASE_PATH/$NAME-386.deb
  package_cloud push $PREFIX/debian/jessie  $RELEASE_PATH/$NAME-386.deb
  package_cloud push $PREFIX/debian/wheezy  $RELEASE_PATH/$NAME-386.deb

else
  make
  make test
fi
