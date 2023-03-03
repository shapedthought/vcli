# User Guide

## API Versions

The default API versions are as follows:

| Product            | Version | API Version | Notes     |
| ------------------ | ------- | ----------- | --------- |
| VBR                | 12.0    | 1.1-rev0    |           |
| Enterprise Manager | 12.0    | -           | No Change |
| VONE               | 12.0    | 1.0-rev2    |           |
| V365               | 7.0     | -           |           |
| VBAWS              | 5.0     | 1.2-rev0    |           |
| VBAZURE            | 4.0     | -           |           |
| VBGCP              | 4.0     | 1.0-rev0    |           |

If you wish to change API version, please update the profile in the profiles.json file.

https://www.veeam.com/documentation-guides-datasheets.html

## Init

To start using the app enter:

    ./vcli.exe init

This will create two files:

- settings.json - contains settings that can be adjusted
- profiles.json - profiles of each of the APIs

If you have set the VCLI_SETTINGS_PATH in the environmental variables before running this command, the files will be located in that directory. Otherwise they will be created in the directory you ran the command.

Note that the directory will not be automatically created for you, it will need to be in place before running the command.

Examples:

Windows

    "C:\User\UserName\.vcli\"

Linux

    "home/veeam/.vcli/"

## Profiles

The profiles.json file contains key information for each of the APIs, these mainly differ in terms of Port, and login URL.

The profiles currently are:

- vbr
- ent_man (Enterprise Manager)
- vb365
- vone
- aws
- azure
- gcp

To see a list of the profiles run

    ./vcli profile --list / -l

To see the current set profiles run

    ./vcli profiles --get / -g

To set a new profile run

    ./vcli profiles --set / -s

To see the details of a profile run

    ./vcli profiles --profile / -p

    {
    "name": "ent_man",
    "headers": {
        "accept": "application/json",
        "Content-type": "application/json",
        "x-api-version": ""
    },
    "url": ":9398/api/sessionMngr/?v=latest",
    "port": "9398",
    "api_version": "",
    "username": "administrator@",
    "address": "192.168.0.123"
    }

## Log in

_UPDATED_

There are now two modes to log in, the first is the "environmental" mode, and the second takes some of the credentials data from the profile.json file aka "creds file" mode.

### Environmental mode

When running the "init" command select N/No to use this mode.

Before logging in you will need to set the following environmental variables:

| Name               | Description                                                                        |
| ------------------ | ---------------------------------------------------------------------------------- |
| VCLI_USERNAME      | The username of the API you are logging into                                       |
| VCLI_PASSWORD      | The password of the API you are logging into                                       |
| VCLI_URL           | The address of the API (without the https:// at the start or the :port at the end) |
| VCLI_SETTINGS_PATH | Optional, sets the location for the settings and configuration files               |

As stated before if you have set the VCLI_SETTINGS_PATH before running "init" the files will be located there. If you set it after then you will need to manually move the files to that location before running further commands.

### Creds file mode

When running the "init" command it will ask if you wish to use the Creds file mode, if yes then the tool will read the username and address of the API from the profiles.json file.

You will need to locate and update the profiles with details before calling any of the APIs.

The password will still need to be set as an environmental variable VCLI_PASSWORD.

Doing this provides faster switching between APIs, but does expose the API username and address in the **clear** in the profiles.json file.

### Logging in

After setting up the credentials using either of the methods above and setting the required Profile, next you can simply run the following:

    ./vcli.exe login

If successful it will save a headers.json file which includes the API key that will be used for future calls.

NOTE: The API key is overwritten on each login so switching between profiles will require you to re-login. This may change in the future.

### Change Modes

Simply locate the settings.json file and update the "credsFileMode" to true or false:

    "credsFileMode":true

## API Commands overview

The tools has also been designed to allow you to output the responses to json and yaml formats. These allow you to then modify these responses using tools such as jq.

However, we have found that pairing vcli with "nu shell" provides an excellent user experience for manipulating API responses.

https://www.nushell.sh/

See the nushell section below.

## Get

With "get" pass the endpoint that you want to get data from after the API version number. The response is always json unless you pass the --yaml flag.

    vcli get jobs

### Example

To get all managed servers from VBR the full endpoint is:

    /api/v1/backupInfrastructure/managedServers

You would pass the following

    vcli get backupInfrastructure/managedServers

## Post / Put

With "post" if the endpoint does not need data sent, then simply enter the end of the URL

    vcli post jobs/57b3baab-6237-41bf-add7-db63d41d984c/start

If the endpoint requires data then use the -f flag with the path to a JSON file.

    vcli post jobs -f job_data.json

With "put" the endpoint must have a payload.

    vcli put jobs -f job_data.json

## Utils

_new in 0.5.0_

The utils command allows you to run a number of utility commands.

The current options are:

- VBR Job JSON Converter
- Check Version

### VBR Job JSON Converter

As the VBR API's GET job object is different from the POST job object, this utility allows you to convert a GET job object to a POST job object.

You will need to get a single job from the VBR API for this to work, it will not work on the "Get All Jobs" endpoint response object (something I might look at later).

### Example

Get a single job from the VBR API

    vcli get jobs/57b3baab-6237-41bf-add7-db63d41d984c > job.json

Run the utils command

    vcli utils

- Select "VBR Job JSON Converter"
- Enter the path to the job json file
- Enter the path to the output file e.g. job_updated.json
- Modify the job json file as required
- POST the job

```
vcli post/jobs -f job_updated.json
```

Note that VBR PUT job json/objects match the GET json/objects, so you can use the same file for PUT.

### Check Version

This command will check the version of the tool against the latest release on GitHub.

## Using with jq

[jq](https://stedolan.github.io/jq/) offers a lightweight way to navigate JSON date, simply pipe out the responses from the API and use JQ to manipulate the data.

Enterprise Manager Example:

     vcli get jobs?format=entity | jq '.Jobs[].JobInfo.BackupJobInfo.Includes.ObjectInJobs[] | .Name, .ObjectInJobId'

Useful tip: use jq "keys" to see the keys in a json object

    vcli get jobs | jq 'keys'

## Using with Nushell

https://www.nushell.sh/

![nushell](./assets/nushell.png)

Is designed to work around structured data in a better way that normal shells do so it is ideal for manipulating data from APIs.

Installation: https://www.nushell.sh/book/installation.html

Personally I use chocolatey.

Once installed you just need to enter the command:

    nu

Then if you have vcli installed you can set the environmental variables like so:

    let-env VCLI_USERNAME = "username"
    let-env VCLI_PASSWORD = "password"
    let-env VCLI_URL = "192.168.0.123"

Then login

    vcli login

As vcli prints json to the screen you can simply pipe the output into nushell

    vcli get jobs | from json

As most of the APIs hold the actual data under a "data" object you will need to "get" that data

    vcli get jobs | from json | get data

Nushell has a huge amount of methods to explore, filter and transform you data. One of my favorite is being about to pipe out to a different format.

    vcli get jobs | from json | get data | to yaml

You can also save it in a different format, though you need to use the --raw flag.

    vcli get jobs | from json | get data | to yaml | save jobs.yaml --raw

### Nu Modules

Nushell has its own module system which means you can define a series of methods which can then be brought into the shell's scope.

https://www.nushell.sh/book/modules.html

For example, create a file called v.nu and add the following:

    # v.nu

    export def vget [url: string] {
        vcli get $url | from json | get data
    }

You can then import the module by entering:

    use v.nu

You can then use that module like so:

    v vget jobs

Which does the same as if you did the longer version shown above.

Using modules means that you can easily create complex queries very easily and be able to recall them when needed.

It is also possible add the environmental variables to the module making it even easier to get going.

    #v.nu

    export-env {
        let-env VCLI_USERNAME = "username"
        let-env VCLI_PASSWORD = "password"
        let-env VCLI_URL = "192.168.0.123"
    }

    export def vlogin [] {
        vcli login
    }

    export def vget [url: string] {
        vcli get $url | from json | get data
    }

Nushell has it's own HTTP get and post options, which could be turned into a specific module for Veeam, however, vcli has been designed to do all that already.

There is also a plugin system that Nushell provides which might be something I look at in the future.

## Tips and Tricks

### Replacing a parameter in a JSON file

There is great tool (written in Rust 🦀) called sd which works like sed and allows you to replace strings in a file using string expressions and regex.

For example, if you wanted to replace the name of a job in a JSON file you could do the following:

Check the name using jq

    cat job.json | jq '.name'

The replace the name using sd

     sd '"name": "Backup Job 2"' '"name": "Backup Job 12"' .\job.json

You can also pipe the vcli output directly into sd to update a parameter.

    vcli get jobs/57b3baab-6237-41bf-add7-db63d41d984c | sd '"name": "Backup Job 2"' '"name": "Backup Job 12"' > job.json

I find this useful to make quick changes to a file without having to open it in a text editor.

sd tool: https://crates.io/crates/sd

Install using chocolatey

    choco install sd-cli

Install via Cargo (Rust package manager)

    cargo install sd
