#bin/bash
# mkdir ./bin
rm -r ./bin/*
mkdir -p ./bin/config && cp ./config/config.yml ./bin/config/

swag init -g /handler/router.go -o ../jim-docs/docs
