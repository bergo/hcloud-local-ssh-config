# HCLOUD LOCAL SSH CONFIG

The tools helps to generate a local ssh config file based on your hetzner cloud server projects.

It's able perform multiple projects and and set default and custom server configuration per project.

The result configuration content can is attached or replaced to a existing config file. It is using generated markers to identify the generated content for a later replacement.

## Install

Build the tool.

    make

Configure the projects by setting up the config.json

    cp config.json.example config.json

Set the projects with name and token in config.json


## Usage

    ./hcloud-local-ssh-config --help
    Usage of ./hcloud-local-ssh-config:
    -config-file string
            configuration file (default "config.json")
    -marker string
            hcloud replacement marker (default "HCLOUD-REPLACE")
    -printonly
            don't write to file print out only
    -ssh-config-file string
            ssh configuration file (default "~/.ssh/config")

## License
MIT - Stefan Berger