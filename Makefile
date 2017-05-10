# build static binary

all:
	## make building faster (for development only)
	# env CGO_ENABLED=0 go install -v

	## ignore vendor dir
	
	@echo moving vendor dir to vendor.mv && echo
	mv -nvi vendor vendor.mv || true
	@echo
	@echo building static binary && echo
	env CGO_ENABLED=0 go build -v -x -ldflags='-s -w' -o markdownd
	@echo
	@echo putting back vendor dir
	mv -nvi vendor.mv vendor
	
