set -o errexit
set -o pipefail
set -o nounset

echo "Working Dir: $PWD"
mkdir -p bin

cd src

LDFLAGS="-w -s"
#GOOS=windows  GOARCH=amd64                go build -ldflags "${LDFLAGS}" -o ../bin/slapperx_win.exe
GOOS=linux    GOARCH=amd64  CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o ../bin/slapperx_linux
#GOOS=darwin   GOARCH=amd64                go build -ldflags "${LDFLAGS}" -o ../bin/slapperx_mac_amd64
#GOOS=darwin   GOARCH=arm64                go build -ldflags "${LDFLAGS}" -o ../bin/slapperx_mac_arm64
