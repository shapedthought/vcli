# veeamcli

Early stages of a Veeam CLI project.

Currently only supports VBR.

To start using the app enter:

    ./veeamcli.exe init

This will create several files:

    settings.json - contains settings that can be adjusted
    profiles.json - profiles of each of the APIs
    creds.yaml - a file to enter the credentials of the API you are using

After filling in the required information in the creds.yaml file you will need to login to the API

    ./veeamcli.exe login

This will login to the API and save a headers.json file which includes the API key that will be used for future calls.

Note: if you are using a self-signed certificate you will need to change the "apiNotSecure" field in the settings.json file to "true".