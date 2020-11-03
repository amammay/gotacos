# Create templates
mkdir -p templates &&
  wget https://raw.githubusercontent.com/pacedotdev/oto/master/otohttp/templates/server.go.plush -q -O ./templates/server.go.plush &&
  wget https://raw.githubusercontent.com/pacedotdev/oto/master/otohttp/templates/client.go.plush -q -O ./templates/client.go.plush

# generate server stub
mkdir -p generated/server

oto -template ./templates/server.go.plush -out ./generated/server/oto-server.gen.go -pkg server ./definitions
gofmt -w ./generated/server/oto-server.gen.go ./generated/server/oto-server.gen.go

# generate client stub

mkdir -p generated/client

oto -template ./templates/client.go.plush -out ./generated/client/oto-client.gen.go -pkg client ./definitions
gofmt -w ./generated/client/oto-client.gen.go ./generated/client/oto-client.gen.go
