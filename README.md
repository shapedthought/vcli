# veeamcli

Early stages of a Veeam CLI project.

NOTE: This is not an official Veeam tool and is provided under the MIT license.

This is project is also in the early stages of development, there maybe future breaking changes.

## What is it?

The veeamcli provides a single interface to work with all Veeam products, these include:

- VBR
- VB365
- VONE - Partial
- VB for Azure - TBA
- VB for AWS -TBA
- VB for GCP -TBA
- VSPC - TBA

It uses each of the products' API to provide a simple-to-use interface for basic day-to-day operations.

Additional functionality includes certain reports that usually require separate scripts.

## Why?

The main aim here is to make using Veeam APIs more accessible by handling many of the
barriers to entry such as authentication.

If you are already a power API user with your own tools, then fantastic, please carry on with what you are doing!

The benefits include:

- simple to install
- simple syntax
- run anywhere
- run on anything

VBR and VB365 have powerful PowerShell cmdlets which I encourage you to use if that is your preference. Veeamcli is not designed as a replacement, just a compliment.

However, products such as VB for AWS/Azure/GCP do not have a command line interface, this is where the veeamcli can really help.

## How to use

Please see the user guide for more information

https://github.com/shapedthought/veeamcli/user_guide.md

## Installing

<b>IMPORTANT The only trusted source of this tool is directly from the release page, or building from source.</b>

<b>Please also check the checksum of the downloaded file before running.</b>

Veeamcli runs in place without needing to be installed on a computer.

It can be added to your system's path to allow system wide access.

To download please go to the releases tab.

windows

    Get-FileHash -Path <file path> -Algorithm SHA256

Mac

    shasum -a 256 <file path>

Linux - distributions may vary

    sha256sum <file path>

## Compiling from source

If you wish to compile the tool from source, please clone this repo, install Golang, and run the following in the root directory.

Windows:

    go build -o veeamcli.exe

Mac/ Linux

    go build -o veeamcli

## Docker 🐋

NOTE: at time of writing there will be <b>no official docker image.</b>

To run veeamcli in an isolated environment you can use a Docker container.

    docker run --rm -it ubuntu bash

    wget <URL of download>

    ./veeamcli init

Persisting the init fils can be done with a bindmount, but note that this does open a potential security hole.

You can of course create your own docker image with veeamcli installed which can be built from source.

Example using local git clone:

    FROM golang:1.18 as build

    WORKDIR /usr/src/app

    COPY go.mod go.sum ./

    RUN go mod download && go mod verify

    COPY . .

    RUN go build -v -o /usr/local/bin/veeamcli

    FROM ubuntu:latest

    WORKDIR /home/veeamcli

    RUN apt-get update && app-get upgrade -y && apt-get install vim -y

    COPY --from=build /usr/local/bin/veeamcli ./

Then exec into the container

    docker run --rm -it txtxx56/veeamcli:0.2 bash

    cd /home/veeamcli

    ./veeamcli init

With the --rm flag the container will deleted immediately after use.

Even when downloading the veeamcli into a Docker container, ensure you check the checksum!

## Why Golang?

The main reason for using Golang, after careful consideration, was that it compiles to a single binary with all depencies included.

Python was a close second, and is a great language, but some of the complexities of dependency management can make it more difficult to get started.

Why not .NET, RUST, etc?

Mainly due to experiance with the language. If you prefer these languages, feel free to take this implementation as an example and build your own.

## Contribution

If you think something is missing, or you think you can make it better, feel free to send us a pull request.

## Issues and comments

If you have any issues or would like to see a feature added please raise an issue.
