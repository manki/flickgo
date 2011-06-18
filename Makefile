include $(GOROOT)/src/Make.inc

TARG=flickr
GOFILES=request.go json.go flickr.go

include $(GOROOT)/src/Make.pkg
