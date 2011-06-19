include $(GOROOT)/src/Make.inc

TARG=flickr
GOFILES=request.go photo.go flickr.go
DEPS=../multipart_writer

include $(GOROOT)/src/Make.pkg
