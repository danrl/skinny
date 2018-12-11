# The Skinny Distributed Lock Service

[![CircleCI](https://circleci.com/gh/danrl/skinny.svg?style=svg)](https://circleci.com/gh/danrl/skinny)
[![codecov](https://codecov.io/gh/danrl/skinny/branch/master/graph/badge.svg?token=u9RZS2ts8s)](https://codecov.io/gh/danrl/skinny)
[![Go Report Card](https://goreportcard.com/badge/github.com/danrl/skinny)](https://goreportcard.com/report/github.com/danrl/skinny)
[![GolangCI](https://golangci.com/badges/github.com/danrl/skinny.svg)](https://golangci.com/r/github.com/danrl/skinny)
[![GoDoc](https://godoc.org/github.com/danrl/skinny?status.svg)](https://godoc.org/github.com/danrl/skinny)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)

![Giraffe](doc/img/giraffe-small.png)


## Welcome

This is the Skinny distributed lock service. It is feature-free and uses a Paxos-like protocol for reaching distributed
consensus. The purpose of this software is to educate its author and interested folks on the broader topic of
distributed systems, reliability, and particularly distributed consensus. The naming was totally not inspired by
Google's wonderful [Chubby lock service for loosely-coupled distributed systems](https://static.googleusercontent.com/media/research.google.com/en//archive/chubby-osdi06.pdf). Or was it?

**Note:** *This is a hobby project of mine. I wanted to learn about distributed consensus and subsequently spread
the knowledge within the Site Reliability Engineering community. This software shall under no circumstances be run in a
production environment. Furthermore, it shall not be trusted in any way. This code is shared for educational purposes
only. Have fun, tinker, learn!*


## Building

* Install build dependencies first
  * Skinny uses the [Mage](https://magefile.org/) build tool. Therefore it needs the `mage` command to be installed.
  * Skinny uses [Protocol Buffers](https://developers.google.com/protocol-buffers/) and [gRPC](https://grpc.io/).
  It requires the `protoc` compiler with the *go* and *grpc* output plugins installed.
  * Some build dependencies can be installed by running `mage builddeps`.
* Download code dependencies by running `mage deps` or `go mod vendor`.
* Build the `skinnyctl` client tool and the `skinnyd` server binaries by running `mage build`.
  You can find the final artifacts in the `./bin/` directory.

Skinny comes with few code dependencies. 


## The Daemon (skinnyd)

A Skinny instance is started by running `skinnyd`, preferably with the `--config` option.

    ./bin/skinnyd --config doc/examples/skinnyd/london.yml

The Skinny instance will add the peers it finds in the configuration file to the quorum.
For each instance in the quorum, a separate `skinnyd` process is started.
Instances are expected to be able to reach out to each other via HTTP/2.
To run multiple instances locally, it is advised to assign each instance its own port and listen on `localhost`.

A typical Skinny configuration file for a Skinny instance in a quorum of five looks like this:

~~~yaml
---
name: london
increment: 1
timeout: 500ms
listen: 0.0.0.0:9000
peers:
- name: oregon
  address: oregon.skinny.cakelie.net:9000
- name: spaulo
  address: spaulo.skinny.cakelie.net:9000
- name: sydney
  address: sydney.skinny.cakelie.net:9000
- name: taiwan
  address: taiwan.skinny.cakelie.net:9000
~~~

All options are required.

| Option            | Description |
| ----------------- | ----------- |
| **Name**          | The name of the Skinny instance. Should be unique within the quorum to avoid confusion. |
| **Increment**     | The number by which the instance increases the round number (ID). Must be unique within the quorum to prevent dueling proposers. |
| **Timeout**       | The timeout for Remote Procedure Calls (RPCs) made to other Skinny instances in the quorum. |
| **Listen**        | The listening address of the Skinny instance. Other instances can connect to this address for RPCs. |
| **Peers**         | The complete list of the *other* instances of the quorum. Should contain an even number of peers. |
| **Peers/Name**    | The name of a peer instance. |
| **Peers/Address** | The address under which a peer instance's RPCs are exposed. |


There must be one configuration file for each Skinny instance in the quorum.
Example configuration files are available in the [`doc/examples`](doc/examples) directory.


## The Client Tool (skinnyctl)

A quorum of Skinny instances is controlled via `skinnyctl`.

The tool implements two APIs:

* The rather simple *control* API. Used for fetching the current status of a quorum.
* The barely more complex *lock* API. This one acquires and releases *locks* on behalf of a *holder*.

To be able to work with a quorum of instances `skinnyctl` needs to know about it. The quorum's connection information is
usually stored in a `quorum.yml` configuration file.

~~~yaml
---
timeout: 5s
instances:
- name: london
  address: london.skinny.cakelie.net:9000
- name: oregon
  address: oregon.skinny.cakelie.net:9000
- name: spaulo
  address: spaulo.skinny.cakelie.net:9000
- name: sydney
  address: sydney.skinny.cakelie.net:9000
- name: taiwan
  address: taiwan.skinny.cakelie.net:9000
~~~

All options are required.

| Option                | Description |
| --------------------- | ----------- |
| **Timeout**           | The timeout for Remote Procedure Calls (RPCs) made to the Skinny instances in the quorum. |
| **Instances**         | The complete list of all instances of the quorum. |
| **Instances/Name**    | The name of an instance. |
| **Instances/Address** | The address under which an instance's RPCs are exposed. |


## Acquiring and Releasing a Lock

**Note:** Locks are always advisory. There is neither a dead-lock detection nor are locks enforced. The holder of a lock is
responsible for releasing the lock after leaving the critical section of an application.

To acquire a lock in behalf of a holder named *Beaver* simply run:

    $ ./bin/skinnyctl acquire "Beaver"
    📡 connecting to london (london.skinny.cakelie.net:9000)
    🔒 acquiring lock
    ✅ success

Once *Beaver* is done accessing the protected resource the lock should be released so that other potential holders can
acquire it.


    $ ./bin/skinnyctl release
    📡 connecting to london (london.skinny.cakelie.net:9000)
    🔓 releasing lock
    ✅ success


### Monitoring Quorum State

A quorum's state can be fetched by issuing a request for status information to every instance in the quorum.

    $ ./bin/skinnyctl status
    NAME     INCREMENT   PROMISED   ID   HOLDER   LAST SEEN
    london   1           1          1    beaver   now
    oregon   2           1          1    beaver   now
    spaulo   3           1          1    beaver   now
    sydney   4           1          1    beaver   now
    taiwan   5           1          1    beaver   now

To continously monitor a quorum's state use the `--watch` option.

    $ ./bin/skinnyctl status --watch

![](doc/img/skinnyctl-status-watch.gif)


## Bonus: Lab Infrastructure via Terraform

Terraform definitions and a *skinny_instance* module are available in the [`doc/terraform`](doc/terraform) directory.
Change the `variables.tf` to your needs and initialize and deploy the environment via:

    $ terraform init
    $ terraform apply

It takes a couple of minutes to fire up all the resources. Be patient.


## Bonus: Lab Software Deployment via Ansible

Ansible playbooks for deploying a Skinny quorum are available in the [`doc/ansible`](doc/ansible) directory.
Adjust the hostnames in the `inventory.yml` file and run the playbook:

    $ ansible-playbook -i inventory.yml site.yml

It takes a while to fetch the source code and build the binaries on each instance.
While you wait, how's the weather today? Well, well...


## Sources and Acknowledgements

* [*Reaching Agreement in the Presence of Faults*](https://lamport.azurewebsites.net/pubs/reaching.pdf), M. Pease, R, Shostak, and L. Lamport
* [*Paxos Made Simple*](https://lamport.azurewebsites.net/pubs/paxos-simple.pdf), L. Lamport
* [*Paxos Agreement - Computerphile*](https://youtu.be/s8JqcZtvnsM), Heidi Howard
  (University of Cambridge Computer Laboratory)
* [*The Paxos Algorithm*](https://youtu.be/d7nAGI_NZPk), Luis Quesada Torres (Google Site Reliability Engineering)
* Giraffe and beaver graphics by OpenClipart contributor
  [Lemmling](https://openclipart.org/user-detail/lemmling)
* Alien graphic by OpenClipart contributor [Anarres](https://openclipart.org/user-detail/anarres)


## License

Copyright 2018 Dan Lüdtke <mail@danrl.com>

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
