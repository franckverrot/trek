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

Here's how to use it:

    λ ./trek -list-jobs
    * example
    * example34

    λ trek -job example34
    * cache34
    * cache56

    λ trek -job example34 -task-group cache56
    * example34.cache56[0]

    λ trek -job example34 -task-group cache56 -allocation 0
    (0) redis5
    (1) redis6

    λ trek -job example34 -task-group cache56 -allocation 0 -task-name redis6
    * Name: redis6
    * Node Name: feynman.local
    * Node IP: 127.0.0.1
    * Driver: docker
            * image: redis:3.2
            * port_map: [map[db:6379]]
    * Dynamic Ports: 24832 (db)


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
