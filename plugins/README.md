This folder contains some plugin examples, each in its directory.

To use the plugins, compile them, then move the .so file to the data directory for your honeytrap install (for example, /home/peter/.honeytrap):

    $ cd plugins/udp-ampl-detector
    $ go build -buildmode=plugin udp-ampl-detector.go # For smaller plugin sizes, append "-ldflags '-s'"
    $ cp udp-ampl-detector.so /home/peter/.honeytrap

To write plugins, write them under "package main" in a separate directory, and make sure to export the correct function(s) - Transform for transforms, Service for services, and so on. For more details, refer to the documentation.