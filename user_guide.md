# User Guide

## Init

To start using the app enter:

    ./veeamcli.exe init

This will create two files:

- settings.json - contains settings that can be adjusted
- profiles.json - profiles of each of the APIs

## Profiles

The profiles.json file contains key information for each of the APIs, these mainly differ in terms of Port, and login URL.

The profiles currently are:

- vbr
- vbm 365
- vone
- aws
- azure
- gcp

To see a list of the profiles run

    ./veeamcli profile --list / -l

To see the current set profiles run

    ./veeamcli profiles --get / -l

To set a new profile run

    ./veeamcli profiles --set / -s <api name>

## Log in

Before logging in you will need to set the following environmental variables:

| Name          | Description                                  |
| ------------- | -------------------------------------------- |
| VCLI_USERNAME | The username of the API you are logging into |
| VCLI_PASSWORD | The password of the API you are logging into |
| VCLI_URL      | The address of the API                       |

After doing this and setting the required Profile, you will need to login to the API:

    ./veeamcli.exe login

Note: if you are using a self-signed certificate you will need to change the "apiNotSecure" field in the settings.json file to "true".

If successful it will save a headers.json file which includes the API key that will be used for future calls. You will need to run the tool in the same directory as these files reside.

## Commands overview

The tools has also been designed to allow you to output the responses to json and yaml formats. These allow you to then modify these responses using tools such as jq.

However, we have found that pairing veeamcli with "nu shell" provides an excellent user experience for manipulating API responses.

https://www.nushell.sh/

## Get

With get pass the endpoint that you want to get data from after the API version number. The response is always json unless you pass the --yaml flag.

### Example

To get all managed servers from VBR the full endpoint is:

    /api/v1/backupInfrastructure/managedServers

You would pass the following

    veeamcli get custom backupInfrastructure/managedServers

    veeamcli get custom backupInfrastructure/managedServers --yaml
