#!/bin/bash

HERE=`pwd`
ARCH=$(uname -m)
ACTION=${1}
declare -a folders=("of-builder" "router" "auth")

if [ "$ARCH" = "armv7l" ] ; then
   ARM_VERSION="armhf"
elif [ "$ARCH" = "aarch64" ] ; then
   ARM_VERSION="arm64"
else
   echo "Not running on ARM. Exiting..."
   exit 1
fi

echo "Executing ${ACTION} action for ${ARM_VERSION} target architecture"

for i in "${folders[@]}"
do
    cd $HERE/$i && make ci-${ARM_VERSION}-${ACTION}
done
