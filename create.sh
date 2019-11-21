#!/usr/bin/env bash
docker rmi -f petrjahoda/rompa_ppa_import_service:"$1"
docker build -t petrjahoda/rompa_ppa_import_service:"$1" .
docker push petrjahoda/rompa_ppa_import_service:"$1"