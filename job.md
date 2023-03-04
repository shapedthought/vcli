# vCLI jobs

### What is it?

vCLI Jobs has been created to make it easier to create jobs.

Normally with the APIs you would need to handle a large job object, for example VBR's job object is:

- type
- id
- name
- description
- isDisabled
- isHighPriority
- virtuaMachines {}
- storage {}
- guestProcessing {}
- schedule {}

Most environments will only need to update the virtual machines element with the virtual machines that need to be backed up. The storage, schedule, and guest processing tend to have less variation.

vCLI jobs allow you to use a base job template which holds the your default values for your jobs.

If you only want to create a job that differs from the base job in the virtual machines then you can create a job file alone (see below for structure).

However, if you want to modify other aspects you will need to create files for that is different from the base file e.g. storage, guest processing, and schedule.

With the job only option you can run the following command against a file containing the "job" object:

```
vcli job create --vbr abc-job.yaml
```

This method should allow you to create jobs 🔥 blazingly fast 🔥 and with a lot less boilerplate to manage, in fact the smallest object you can use is:

```yaml
type: Backup
name: my job
description: Created by VCLI!!!.
isDisabled: false
virtualMachines:
  includes:
    - type: VirtualMachine
      hostName: your-vmware-host
      name: your-vm-name
      objectId: vm-0001
```

If you want to modify other aspects, the you run the following command against a folder containing the files:

```
vcli job create --vbr -f /path/to/job/folder
```

## Templates

The template command will take a specified job ID and convert into the following files:

- abc-job.yaml
- storage.yaml
- guest-processing.yaml
- schedule.yaml
- job-template.yaml

The command will automatically convert the GET job object to a POST job object if the objects are different like in the case of VBR.

```
vcli job template --vbr jobs/57b3baab-6237-41bf-add7-db63d41d984c
```

You will then need to place the base-job.yaml file in the settings folder. This also requires that you are using the VCLI_SETTINGS_PATH environment variable (see user guide).

```
cp base-job.yaml /path/to/settings/folder/base-job.yaml
```

You can then use the other template files as the bases for your jobs.

## Job File Structure

The job file contains the following elements:

- name
- description
- isDisabled
- isHighPriority
- virtualMachines {}

The structure of the virtual machines element needs to follow the structure laid out here:

https://helpcenter.veeam.com/docs/backup/vbr_rest/reference/vbr-rest-v1-1-rev0.html?ver=120#tag/Jobs/operation/CreateJob

However, the job template will likely give you most of what you need.

## Job Folder Structure

The job folder structure is as follows:

```
/path/to/job/folder
├── abc-job.yaml
├── guest-processing.yaml
├── schedule.yaml
└── storage.yaml
```

Remember, you only need to include the files for the elements that differ from the base job template.

Note: The job file must contain the word "job" in the file name.

## How it works

1. vCLI will first load the base job template.
2. It will then load the job related file(s)
   - For the job only option it will only load the job file.
   - For the job folder option (-f) it will then load all the files in the job folder.
3. It will then merge the file or files together with the base job template.
4. Finally it will post the job to the specified endpoint.

## Supported Job Types

Current it is only VBR, more will come in the future.
