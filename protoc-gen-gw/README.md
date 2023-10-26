### install 
$ go install

### gen-go plugin version 
1.31.0

### generate pb
$ protoc --go_out=. Test.proto

### generate gw
$ protoc --gw_out . Test.proto