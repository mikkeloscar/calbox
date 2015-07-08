GO=go

all: calbox

calbox: clean
	$(GO) build

install:
	# bin
	install -Dm755 calbox $(DESTDIR)/usr/bin/calbox
	# service
	install -d $(DESTDIR)/usr/lib/systemd/system/
	install -m644 contrib/calbox.service $(DESTDIR)/usr/lib/systemd/system/

clean:
	-@rm -f calbox
