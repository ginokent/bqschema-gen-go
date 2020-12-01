# bqtableschema

BigQuery table schema struct generator

[![main](https://github.com/djeeno/bqtableschema/workflows/main/badge.svg)](https://github.com/djeeno/bqtableschema/tree/main)
[![codecov](https://codecov.io/gh/djeeno/bqtableschema/branch/main/graph/badge.svg)](https://codecov.io/gh/djeeno/bqtableschema)

## generate

```console
$ cd /path/to/your/golang-project-repository

$ ## Create a directory where bqtableschema will generate the code.
$ mkdir -p bqtableschema
$ cd bqtableschema

$ ## Set the required environment variables.
$ export GOOGLE_APPLICATION_CREDENTIALS=/path/to/serviceaccount/keyfile.json
$ export GCLOUD_PROJECT_ID=bigquery-public-data  ## ref. https://console.cloud.google.com/bigquery?p=bigquery-public-data&page=project
$ export BIGQUERY_DATASET=hacker_news            ## ref. https://console.cloud.google.com/bigquery?p=bigquery-public-data&d=hacker_news&page=dataset

$ ## generate
$ go run github.com/djeeno/bqtableschema
```
