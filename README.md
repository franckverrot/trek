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

    ./trek -ui=false <task name>

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