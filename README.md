# go-importers

## Installation

```sh
go install github.com/appilon/go-importers
```

## Usage

```sh
GITHUB_PERSONAL_TOKEN=... go-importers > report.json

# JQ can be used to filter output, e.g. if you only want projects w/ >90 stars
cat report.json | jq -r '.[] | select(.stars > 90) | .'
```
