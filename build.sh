#!/bin/bash
set -e

# build static binary

## make building faster (for development only)
# env CGO_ENABLED=0 go install -v

## ignore vendor dir
build(){	
	echo moving vendor dir to vendor.mv && echo
	mv -nvi vendor vendor.mv || true
	echo
	echo building static binary && echo
	env CGO_ENABLED=0 go build -v -x -ldflags='-s -w' -o markdownd
	echo
	echo putting back vendor dir
	mv -nvi vendor.mv vendor
}

if [ -z "$@" ]; then 
build && echo "successfully built static binary: markdownd" && \
exit 0
fi

if [ "package" == "$1" ]; then
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
        cp LICENSE $i/; 
        cp -avx docs $i/; 
        mkdir -p $i/src; 
        cp *.go $i/src/; 
        cp SHA256.txt $i/;
        zip -r $i.zip $i/;
        tar czf $i.tar.gz $i; 
        rm -rvf $i;
done
exit 0
fi

echo '
build.sh usage: run with no arguments to build static binary
./build.sh package to build release packages
'