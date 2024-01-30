openapi-generator-cli generate -i api/openapi.yml -g go-server -o internal/ --additional-properties=packageName=api,router=chi,sourceFolder=api,outputAsLibrary=true
