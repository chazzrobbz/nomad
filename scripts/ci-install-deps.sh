#!/usr/bin/env bash

echo "Installing dependencies for goos:$1 goarch:$2"

export DEBIAN_FRONTEND=noninteractive

# Update and ensure we have apt-add-repository
apt-get update
apt-get install -y software-properties-common

# Add i386 architecture (for libraries)
dpkg --add-architecture i386

# Update with i386, Go and Docker
apt-get update

# Install Core build utilities for Linux
apt-get install -y \
	build-essential \
	git \
	libc6-dev-i386 \
	libpcre3-dev \
	linux-libc-dev:i386 \
	pkg-config \
	zip \
	curl \
	jq \
	tree \
	unzip \
	wget

if [[ $1 == "linux" && $2 == "386" ]]; then
    apt-get install gcc-multilib 

# Install ARM build utilities
apt-get install -y \
	binutils-aarch64-linux-gnu \
	binutils-arm-linux-gnueabihf \
	gcc-5-aarch64-linux-gnu \
	gcc-5-arm-linux-gnueabihf \
	gcc-5-multilib-arm-linux-gnueabihf

# Install Windows build utilities
apt-get install -y \
	binutils-mingw-w64 \
	gcc-mingw-w64

# Ensure everything is up to date
apt-get upgrade -y
