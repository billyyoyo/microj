### install 
$ go install

### generate pb
$ protoc --go_out=plugins=grpc:. Test.proto

### generate gw
$ protoc --gw_out . Test.proto