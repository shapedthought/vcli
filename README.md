# vcli

![nu_demo](./assets/main.png)
**Formatting provided by [Nushell.](https://www.nushell.sh/)** vcli only gets the data! See [User guide](https://github.com/shapedthought/vcli/blob/master/user_guide.md) for more information.

NOTE: This is not an official Veeam tool and is provided under the MIT license.

## What is it?

The vcli provides a single interface to get information from Veeam products in an easy way.

These include:

- VBR
- VB365
- VONE
- VB for Azure
- VB for AWS
- VB for GCP

Note that the Enterprise Manager API is currently not supported.

You can also add new endpoints by updating the profiles.json file, but it has been set up for OAuth only (currently).

## Why?

The main aim here is to make using Veeam APIs more accessible by handling many of the barriers to entry such as authentication.

If you are already a power API user with your own tools, then fantastic, please carry on with what you are doing!

The benefits include:

- simple to install
- simple syntax
- run anywhere
- run on anything

VBR and VB365 have powerful PowerShell cmdlets which I encourage you to use if that is your preference. vcli is not designed as a replacement, just a compliment.

However, products such as VB for AWS/Azure/GCP do not have a command line interface, this is where the vcli can really help.

## Why reporting only?

The aim of the project was to keep it light-weight and free of the potential harm which POST and PUTS could cause.

However, the tool does provide a convenient way to login and get the API key, so it could be used with other tools to do modifications.

## How to use

Please see the user guide for more information

https://github.com/shapedthought/vcli/blob/master/user_guide.md

## Installing

<b>IMPORTANT The only trusted source of this tool is directly from the release page, or building from source.</b>

<b>Please also check the checksum of the downloaded file before running.</b>

vcli runs in place without needing to be installed on a computer.

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

    go build -o vcli.exe

Mac/ Linux

    go build -o vcli

## Docker 🐋

NOTE: at time of writing there will be <b>no official docker image.</b>

To run vcli in an isolated environment you can use a Docker container.

    docker run --rm -it ubuntu bash

    wget <URL of download>

    ./vcli init

Persisting the init fils can be done with a bindmount, but note that this does open a potential security hole.

You can of course create your own docker image with vcli installed which can be built from source.

Example using local git clone:

    FROM golang:1.18 as build

    WORKDIR /usr/src/app

    COPY go.mod go.sum ./

    RUN go mod download && go mod verify

    COPY . .

    RUN go build -v -o /usr/local/bin/vcli

    FROM ubuntu:latest

    WORKDIR /home/vcli

    RUN apt-get update && app-get upgrade -y && apt-get install vim -y

    COPY --from=build /usr/local/bin/vcli ./

Then exec into the container

    docker run --rm -it YOUR_ACCOUNT/vcli:0.2 bash

    cd /home/vcli

    ./vcli init

With the --rm flag the container will deleted immediately after use.

Even when downloading the vcli into a Docker container, ensure you check the checksum!

## Why Golang?

The main reason for using Golang, after careful consideration, was that it compiles to a single binary with all decencies included.

Python was a close second, and is a great language, but some of the complexities of dependency management can make it more difficult to get started.

If anyone knows me I love RUST, but for this decided that Go was a better choice.

If you prefer other languages, feel free to take this implementation as an example and build your own.

## Contribution

If you think something is missing, or you think you can make it better, feel free to send me a pull request.

## Issues and comments

If you have any issues or would like to see a feature added please raise an issue.
