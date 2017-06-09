#!/bin/bash
set -e
set -x
GRN="\e[32m"
RED="\e[31m"
RST="\e[0m"

# check for recent go, and grab includes
check_deps(){
	# need go
	if [ -z "$(which go)" ]; then
		echo need latest go, grab from https://golang.org
		exit 111
	fi

	# check go version, warn if out dated
	GOVERSION=$(go version)
	if [  -z "$(echo "$GOVERSION" | grep "go1.8")" ]; then
		printf 'Using unsupported version of go compiler: '
		GOVERSION="$RED""$GOVERSION""$RST"
	else
		GOVERSION="$GRN""$GOVERSION""$RST"
	fi
	printf "$GOVERSION\n"

	# get latest deps if they dont exist
	echo "getting dependencies if they don't exist in "'$GOPATH'
	go get -d -v .
	}


# build static binary

## make building faster (for development only)
# env CGO_ENABLED=0 go install -v


build(){	

	if [ -d "vendor" ]; then
		## ignore vendor dir
		echo moving vendor dir to vendor.mv && echo
		mv -nvi vendor vendor.mv
	fi
	check_deps
	echo building static binary
	time env CGO_ENABLED=0 go build -x -v -ldflags='-s -w' -o markdownd
	EXITCODE=$?
	FILESIZE=$(ls -sh markdownd)
	printf "filesize: $FILESIZE\n"

	if [ -x "$(which file)" ]; then
		file markdownd
	fi
	if [ -x "$(which sha256sum)" ]; then
		printf "calculating sha256... "
		sha256sum markdownd
	fi
	if [ -d "vendor.mv" ]; then
		echo putting back vendor dir
		mv -nvi vendor.mv vendor
	fi
	echo exit $EXITCODE
	exit $EXITCODE
}

if [ -z "$@" ]; then 
build && printf "$GRN**** SUCCESS ****$RST\n" && \
exit 0
fi

if [ "package" == "$1" ]; then

set -x
set -e
# make sure we are in go package
if [ -z "$(ls *go)" ]; then
        echo not a go package
        exit 111
fi

# grab make.go and make executable: https://github.com/aerth/make.go/

if [ ! -x "$(which make.go)" ]; then
	echo grab make.go and make executable: https://github.com/aerth/make.go/
	exit 111
fi
env CGO_ENABLED=0 make.go -o pkg/ -all -v
RWD=$(pwd)
cd pkg && sha256sum * > ../SHA256.txt; cd $RWD;

# package each file in 'pkg' dir
for i in $(ls pkg); do
        echo Packaging $i;
        mkdir -p $i;
        cp pkg/$i $i/$i;
        cp README.md $i/;
        cp CHANGELOG.md $i/;
        cp LICENSE $i/;
        cp -avx docs $i/;
        cp -avx theme $i/;
        mkdir -p $i/src; 
        cp *.go $i/src/;
        cp build.sh $i/src/;
        cp SHA256.txt $i/;
        zip -r $i.zip $i/;
        tar czvf $i.tar.gz $i;
        rm -rvf $i;
done
exit 0
fi

echo '
build.sh usage: run with no arguments to build static binary
./build.sh package to build release packages
'
