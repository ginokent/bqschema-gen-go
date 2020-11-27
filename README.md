# bqtableschema
BigQuery table schema struct generator

![.github/workflows/ci.yaml](https://github.com/djeeno/bqtableschema/workflows/.github/workflows/ci.yaml/badge.svg)

## generate
```console
$ cd /path/to/your/repository

$ export GOOGLE_APPLICATION_CREDENTIALS=/path/to/serviceaccount/keyfile.json

$ export GCLOUD_PROJECT_ID=bigquery-public-data  ## ref. https://console.cloud.google.com/bigquery?p=bigquery-public-data&page=project
$ export BIGQUERY_DATASET=hacker_news            ## ref. https://console.cloud.google.com/bigquery?p=bigquery-public-data&d=hacker_news&page=dataset

$ go run github.com/djeeno/bqtableschema
```
