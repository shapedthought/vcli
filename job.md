# vCLI jobs

Note that this is feature is still in beta and may change in the future.

It also only works with VBR at the moment.

### What is it?

vCLI Jobs has been created to make it easier to create jobs.

Normally with the APIs you would need to handle a large job object, for example VBR's job object is usually around 300 lines long. This can be a lot of boilerplate to manage.

The structure of a VBR job object is as follows:

- type
- id
- name
- description
- isDisabled
- isHighPriority
- virtualMachines {}
- storage {}
- guestProcessing {}
- schedule {}

Most environments will only need to update the virtual machines element with the virtual machines that need to be backed up. The storage, schedule, and guest processing tend to have less variation.

vCLI jobs allow you to use a base job template which holds the your default values for your jobs.

If you only want to create a job that differs from the base job in the virtual machines then you can create a job file alone (see below for structure).

However, if you want to modify other aspects you will need to create files for what is different from the base file (storage, guest processing, and schedule).

With the job only option you can run the following command against a file containing the "job" object:

```
vcli job create abc-job.yaml
```

This method should allow you to create jobs 🔥 blazingly fast 🔥 with the smallest job object being the following:

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

If you want to modify other aspects, the you run the following command against a folder containing template files:

```
vcli job create -f \path\to\job\folder
```

See Job Folder Structure below for more information.

## Side note - YAML

YAML was selected as the file format as it is a human readable format and is easy to create and modify.

## Creating Templates

The template command will take a specified job ID and convert into the following template files:

- abc-job.yaml
- storage.yaml
- guest-processing.yaml
- schedule.yaml
- job-template.yaml

The command will automatically convert the GET job object to a POST job object if the objects are different like in the case of VBR.

```windows
vcli job template 57b3baab-6237-41bf-add7-db63d41d984c
```

You will then need to place the job-template.yaml file in the settings folder. This also requires that you are using the VCLI_SETTINGS_PATH environment variable (see main user guide).

```windows
cp job-template.yaml .\path\to\settings\folder\job-template.yaml
```

You can then use the other template files as the bases for your jobs.

## The Job File Structure

The job file contains the following elements:

- name
- description
- isDisabled
- isHighPriority
- virtualMachines {}

It is mixture of the base job object and the virtual machines element.

The structure of the virtual machines element needs to follow the structure laid out here:

https://helpcenter.veeam.com/docs/backup/vbr_rest/reference/vbr-rest-v1-1-rev0.html?ver=120#tag/Jobs/operation/CreateJob

However, the job template will likely give you most of what you need.

## The Job Folder Structure

The job folder structure is as follows:

```
/path/to/job/folder
├── abc-job.yaml
├── guest-processing.yaml
├── schedule.yaml
└── storage.yaml
```

Remember, you only need to include the files for the elements that differ from the base job template.

The job file must contain the word "job" in the file name. The other files have to keep the names that they were created with e.g. storage, schedule etc.

## How it works

1. vCLI will first load the base job template.
2. It will then load the job related file(s)
   - For the job only option it will only load the job file.
   - For the job folder option (-f) it will then load all the files in the job folder.
3. It will then merge the file or files together with the base job template.
4. Finally it will post the job to the specified endpoint.
