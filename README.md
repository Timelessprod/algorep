# ALGOREP Project

We choose to work on the second project : Distributed job queue system.
Full subject is available in the [subject.pdf](subject.pdf) file.

We choose to develop the project in Go language.

## A) Goal of the project

We wish to set up a client/server system with a mechanism to control or inject
faults into the system. control or inject faults into the system. The general
idea is the following: clients submit jobs to be executed to to be executed to
the servers. These servers then want to agree on the order in which they will
execute these jobs, and which node will be these jobs, and which node will be
responsible for its execution. Once agreed, the jobs can be executed on their
respective their respective nodes, and the result of each job can be retrieved
by the client at the end.

## B) Our group

| Name             | Email (@epita.fr)         | Github account |
| ---------------- | ------------------------- | -------------- |
| Corentin Duchêne | corentin.duchene          | `Nigiva`       |
| Adrien Merat     | adrien.merat              | `Timelessprod` |
| Théo Perinet     | theo.perinet              | `TheoPeri`     |
| Henri Jamet      | henri.jamet               | `hjamet`       |
| Hao Ye           | hao.ye                    | `Enjoyshi`

## C) Running our project

The project require Go version 1.18 or above. To install it, please run the
following commands:
```bash
curl -OL https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

You can check Go versions on the [go.dev](https://go.dev/dl/) website.

Then you can simply clone the repository and run `make` at its root.

## D) Progression

Current advancements on the project, regarding completed steps :

![](https://geps.dev/progress/0)