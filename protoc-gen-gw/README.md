### install 
$ go install

### gen-go plugin version 
1.31.0

### generate pb
$ protoc --go_out=. Test.proto

### generate grpc.pb
$ protoc --go-grpc_out=. Test.proto

### generate gw
$ protoc --gw_out=. Test.proto

### you can write multi plugins in one command