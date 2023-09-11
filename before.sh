#bin/bash
# mkdir ./bin
echo "删除已编译文件"
rm -rf ./bin/*

echo "复制配置文件"
mkdir -p ./bin/config && cp ./config/config.yml ./bin/config/

echo "生成swag文档"
swag init -g /handler/router.go -o ../jim-docs/docs
