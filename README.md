# ALGOREP Project

We choose to work on the second project : **Distributed job queue system.**
Full subject is available in the [subject.pdf](subject.pdf) file.

We choose to develop the project in **Go language**.

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

### C.1) Requirements

⚠️ **The project requires Go version 1.18 or above.** 

To install Go version 1.19.3, please run the
following commands:
```bash
curl -OL https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```
You can check Go versions on the [go.dev](https://go.dev/dl/) website. If you
use another shell than bash, please adapt the last line in consequence.

### C.2) Import, build and run the project

Finally, you can simply : 
* clone the repository 
* run `make` at the root of the project
  
### C.3) How to use the project
When you start the project, you arrive directly on a REPL console. This console allows you to control the cluster, submit jobs and check the status of the jobs.

We provide 8 commands :
- `SPEED (low|medium|high) <node number>` : change the speed of a node. For example: `SPEED high 2` will change the speed of node 2 to high.
- `CRASH <node number>` : crash a node. For example: `CRASH 2` will crash node 2.
- `RECOVER <node number>` : recover a crashed node. For example: `RECOVER 2` will recover node 2.
- `START` : start the cluster. You can use this command only once.
- `SUBMIT <job file>` : submit a job to the cluster. The cluster must be STARTed before. For example: `SUBMIT path/job.cpp` will submit the job described in the file `job.cpp`.
- `STATUS [<job reference>]` : display the status of the cluster or of a specific job. For example: `STATUS` will display the status of the cluster. `STATUS 1-2` will display the status of the job with reference `1-2`.
- `STOP` : stop the cluster. This command will kill the program.
- `HELP` : display this message.

We provide examples of more or less complex jobs in the folder [`examples`](./examples). These jobs end with the extension `.cpp`.

We also provide pre-built scenarios that launch the orders by themselves. To use them, you just have to write `bash example/senario.sh | make`. All scenarios are in [`examples`](./examples) and the files end with the extension `.sh`. For example: 
* `bash examples/scenario-basic-submit.sh | make`
* `bash examples/scenario-multi-submit-with-crash.sh | make`

### C.4) Change the cluster configuration
All the variables defining the shape of the cluster are gathered in a `Config` object that you will find here: [`pkg/core/config.go`](pkg/core/config.go). This will allow you to change the number of Scheduler Node, Worker Node, channel buffer size, timeout duration, number of retries, ...

For the sake of simplicity, we have not implemented several clients in the form of several terminals, but it can be done very well. In any case, the commands sent by the clients will be ordered in a queue which is the channel of the requests to the Leader Scheduler Node.

### C.5) Visualize the cluster
To view the status of jobs, we provided the command `STATUS` and `STATUS <job reference>`. 

To have a more advanced visualization, at each project launch, each Scheduler Node will (over)write a file in the `state` folder listing the contents of its main variables and entries. If you started a cluster of 5 Scheduler Nodes and ran the `START` command, then you will end up with 5 files of the form `<node id>.node` as `0.node`. These files have the following form and update in real time: 
```
---  Scheduler - 0  ---
>>> State:  Follower
>>> IsCrashed:  false
>>> CurrentTerm:  1
>>> LeaderId:  4
>>> VotedFor:  4
>>> ElectionTimeout:  248.498081ms
>>> VoteCount:  0
>>> CommitIndex:  2
>>> MatchIndex:  [0 0 0 0 0]
>>> NextIndex:  [1 1 1 1 1]
### Log ###
[1] Job 1-1 | Worker 0 | WAITING
[2] Job 1-1 | Worker 0 | DONE
----------------
```

💡 To help you view these files, we provide a program [`./show_node_state.sh`](show_node_state.sh) written in shell which shows in each column the contents of the files and which are updated every second. You can run it like this: `bash show_node_state.sh`

To analyze the behavior of the project, we set up a logger system. By default, we have left the log level at `DEBUG` so the file can quickly become large. This file is deleted before each launch by the program itself.

## D) Progression

Current advancements on the project, regarding completed steps :

![](https://geps.dev/progress/100)

**All mandatory** features have been **implemented**.
**Bonus** features have **not** been **implemented**.
