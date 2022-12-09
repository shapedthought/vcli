# User Guide

## Init

To start using the app enter:

    ./veeamcli.exe init

This will create several files:

- settings.json - contains settings that can be adjusted
- profiles.json - profiles of each of the APIs
- creds.yaml - a file to enter the credentials of the API you are using

## Profiles

The profiles.json file contains key information for each of the APIs, these mainly differ in terms of Port, and login URL.

The profiles currently are:

- vbr
- vbm365
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

After filling in the required information in the creds.yaml file, and setting the required Profile, you will need to login to the API:

    ./veeamcli.exe login

Note: if you are using a self-signed certificate you will need to change the "apiNotSecure" field in the settings.json file to "true".

If successful it will save a headers.json file which includes the API key that will be used for future calls. You will need to run the tool in the same directory as these files reside.

To logout out of the API:

    ./veeamcli.exe login --logout

## Commands overview

The veeamcli has certain curated commands that provide output in a tabula format for ease.

The tools has also been designed to allow you to output the responses to json and yaml formats. These allow you to then modify these responses using tools such as jq.

However, we have found that pairing veeamcli with "nu shell" provides an excellent user experience for manipulating API responses.

https://www.nushell.sh/

## Flags

The tool is arranged into several key flags

| flag          | Description                                              |
| ------------- | -------------------------------------------------------- |
| get           | gets data from the API                                   |
| create        | maps to a POST command, and creates a new resource       |
| modify        | maps to a PUT command, and modifies an existing resource |
| start         | starts an action                                         |
| stop          | stops stops an action                                    |
| _verb_ custom | a custom endpoint that isn't current implemented         |

## Get

The get command will pull specific data from the API.

    ./veeamcli get jobs

It also has sub-commands that transform the data into either json, or yaml.

    ./veeamcli get jobs --json / --yaml

There is also a save commands which saves the data to a yaml file.

    ./veeamcli get jobs --save / -s

## Create

The create flag maps to a POST request to create new resources.

As the object required to make these requests tend to be too large to enter directly
into the commandline, there is a required --file / -f flag which takes a yaml file.

    ./veeamcli create job -f job.yaml

Yaml was selected as it provides a more human readable syntax.

As with other tools the advice is to use a get request to create a template, which can be modified to be sent to the API to create a new resource.

## Modify

Modify is almost the same as Create but instead of a POST it uses a PUT request.

    ./veeamcli modify job -f job.yaml

## Start

Start will start an action, for example a job. It is again a POST request, but made into a convenient command.

    ./veeamcli start job --name <job name> / --id <job id>

The name or id flags need to be passed with this command.

## Stop

The same start, but stops specific action.

    ./veeamcli stop job --name <job name> / --id <job id>

## Custom

Custom is where curated command is not available, you can pass any command into the cli, by using the documented endpoint online. The response is always json unless you pass the --yaml flag.

To get all managed servers from VBR the full endpoint is

    /api/v1/backupInfrastructure/managedServers

You would pass the following

    veeamcli get custom backupInfrastructure/managedServers

    veeamcli get custom backupInfrastructure/managedServers --yaml

## Why no delete?

The delete method could cause harm to a system if used incorrectly, because of this it was deemed too risky to add to the project.
