$Env:GOOS = "windows"
$Env:GOARCH = "amd64"
$Env:CGO_ENABLED = 0

go build -o ./build/Service-$Env:GOOS-$Env:GOARCH.exe ./main.go