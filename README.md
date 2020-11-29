# Octopus Energy API

[![PkgGoDev](https://pkg.go.dev/badge/github.com/FileGo/octopusenergyapi/)](https://pkg.go.dev/github.com/FileGo/octopusenergyapi/)
[![Go Report Card](https://goreportcard.com/badge/github.com/FileGo/octopusenergyapi)](https://goreportcard.com/report/github.com/FileGo/octopusenergyapi)
![tests](https://github.com/FileGo/octopusenergyapi/workflows/tests/badge.svg)
![build](https://github.com/FileGo/octopusenergyapi/workflows/build/badge.svg)


This package provides an interface to [Octopus Energy's](https://octopus.energy/) [API](https://developer.octopus.energy/docs/api/), which provides a lot of useful data, including viewing half-hourly consumption of an electricity or gas meter.

An API key from Octopus Energy is required, as mentioned [here](https://developer.octopus.energy/docs/api/#authentication). This can be obtained from your online dashboard.

## Usage

```
client, err := octopusenergyapi.NewClient("{API_KEY}", http.DefaultClient)
if err != nil {
    log.Fatal(err)
}

mpoint, err := client.GetMeterPoint(mpan)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("MPAN: %s\nProfile class: %d (%s)\n",
    mpoint.MPAN, mpoint.ProfileClass, octopusenergyapi.PCs[mpoint.ProfileClass])
fmt.Printf("GSP: %s (%s)\n",
    mpoint.GSP.GSPGroupID, mpoint.GSP.Name)
```

If you find bug or would like to see some additional functionality, please raise an issue. PRs are also more than welcome.