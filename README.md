# Trek

(*Alpha version of a CLI/ncurses explorer for Nomad clusters.*)

Trek is an interface to explore Nomad clusters.

![In Action](https://raw.githubusercontent.com/franckverrot/trek/master/assets/jan-15-screenshot.png)


## SETUP

### Binary Release

Get to revisions, and download a binary.

### From Source

    git clone https://github.com/franckverrot/trek.git
    cd trek
    make trek


## USAGE


*TL;DR* Start `./trek -help` to get the usage prompt.


### CLI

The CLI can be used without a UI. This allows scripting to access IP, ports,
and other info exposed by Nomad.

#### Options

Here's a list of options available:

<a name="nomad-address"></a>
* `nomad-address`: address of the nomad cluster

<a name="list-jobs"></a>
* `list-jobs`: list jobs running on the cluster

```
λ ./trek -list-jobs
* example
* example34
```

<a name="job"></a>
* `job`: select a specific job

```
λ trek -job example34
* cache34
* cache56
```

<a name="task-group"></a>
* `task-group`: select a specific task group

```
λ trek -job example34 -task-group cache56
* example34.cache56[0]
```

<a name="allocation"></a>
* `allocation`: select a specific allocation number

```
λ trek -job example34 -task-group cache56 -allocation 0
(0) redis5
(1) redis6
```

<a name="task-name"></a>
* `task-name`: select a specific task name

```
λ trek -job example34 -task-group cache56 -allocation 0 -task-name redis6
* Name: redis6
* Node Name: feynman.local
* Node IP: 127.0.0.1
* Driver: docker
        * image: redis:3.2
        * port_map: [map[db:6379]]
* Dynamic Ports: 24832 (db)
```

<a name="display-format"></a>
* `display-format`: Use the [Go templating language][go-templating] to format output when describing a specific task
  * Available data:
    * `IP`: no onto which we're running the selected task
    * `Network`: network information (like ports)
    * `Environment`: environment variables provided to the task
  * Available functions:
    * `{{Debug <x>}}` : show raw representation of the data `<x>`
    * `{{DebugAll}}` : show raw representation of everything provided to the template
  * Examples:

```
λ trek -job example34 -task-group cache56 -allocation 0 -task-name redis6 -display-format "{{DebugAll}}"
DEBUG ALL: {IP:127.0.0.1 Network:{Ports:map[db:{Value:23109 Reserved:false}]} Environment:map[FOO_BAR:{Value:baz_bat}]}

λ trek -job example34 -task-group cache56 -allocation 0 -task-name redis6 -display-format "{{Debug .Environment}}"
DEBUG: map[FOO_BAR:{Value:baz_bat}]

λ trek -job example34 -task-group cache56 -allocation 0 -task redis6 -display-format "{{range .Network.Ports}}{{$.IP}}:{{.Value}}{{println}}{{end}}"
127.0.0.1:31478
127.0.0.1:25142
```



### ncurses UI

    ./trek -ui=true


### Trek Configuration File

#### Example

```
{ "Environments" : [ { "Name" : "development" , "Address" : "http://127.0.0.1:4646" }
                   ]
}
```

#### Options

  * `Environments`: List of environments (given a name and address) Trek can connect to



## Todo

* [ ] Make it easy to SSH into a node
* [ ] Better UI
* [ ] More options


## Note on Patches/Pull Requests

* Fork the project.
* Make your feature addition or bug fix.
* Add tests for it. This is important so I don't break it in a
  future version unintentionally.
* Commit.
* Send me a pull request. Bonus points for topic branches.


## Copyright

Copyright (c) 2019 Franck Verrot. MIT LICENSE. See LICENSE for details.


[go-templating]: https://golang.org/pkg/text/template/