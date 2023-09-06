#bin/bash
# mkdir ./bin
rm -r ./bin/*
cp ./config/config.yml ./bin/

swag init -g /handler/router.go -o ../jim-docs/docs
